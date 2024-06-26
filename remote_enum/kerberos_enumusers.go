package remote_enum

import (
	"bufio"
	"context"
	"divine-dragon/transport"
	"divine-dragon/util"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type KerberosEnumUsersModule struct {
	domain       string
	dc           string
	verbose      bool
	safemode     bool
	downgrade    bool
	usernamelist string

	kSession transport.KerberosSession
	kOptions transport.KerberosSessionOptions

	ctx       context.Context
	cancel    context.CancelFunc
	threads   int
	delay     int
	counter   int32
	successes int32

	logger util.Logger
}

func NewKerberosEnumUsersModule(domainOpt string, dcOpt string, verboseOpt bool, safemodeOpt bool,
	downgradeOpt bool, usernamelistOpt string,
	logFileName string, threadsOpt int, delayOpt int) *KerberosEnumUsersModule {
	keum := KerberosEnumUsersModule{domain: domainOpt,
		dc:           dcOpt,
		verbose:      verboseOpt,
		safemode:     safemodeOpt,
		downgrade:    downgradeOpt,
		usernamelist: usernamelistOpt}
	keum.ctx, keum.cancel = context.WithCancel(context.Background())
	keum.logger = util.KerberosEnumUsersLogger(verboseOpt, logFileName)
	keum.delay = delayOpt
	if delayOpt != 0 {
		keum.threads = 1
		keum.logger.Log.Infof("Delay set. Using single thread and delaying %dms between attempts\n", keum.delay)
	} else {
		keum.threads = threadsOpt
	}
	keum.setupSession()
	return &keum
}

func (keum *KerberosEnumUsersModule) setupSession() {
	keum.kOptions = transport.KerberosSessionOptions{
		Domain:           keum.domain,
		DomainController: keum.dc,
		Verbose:          keum.verbose,
		SafeMode:         keum.safemode,
		Downgrade:        keum.downgrade,
	}
	k, err := transport.NewKerberosSession(keum.kOptions)
	if err != nil {
		keum.logger.Log.Error(err)
	}
	keum.kSession = k

	keum.logger.Log.Info("Using KDC(s):")
	for _, v := range keum.kSession.Kdcs {
		keum.logger.Log.Infof("\t%s\n", v)
	}
}

func (keum *KerberosEnumUsersModule) Run() {
	usersChan := make(chan string, keum.threads)
	defer keum.cancel()

	var wg sync.WaitGroup
	wg.Add(keum.threads)

	var scanner *bufio.Scanner
	if keum.usernamelist != "-" {
		file, err := os.Open(keum.usernamelist)
		if err != nil {
			// logger.Log.Error(err.Error())
			return
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for i := 0; i < keum.threads; i++ {
		go keum.makeKerberosEnumUsersWorker(keum.ctx, usersChan, &wg)
	}

	start := time.Now()

Scan:
	for scanner.Scan() {
		select {
		case <-keum.ctx.Done():
			break Scan
		default:
			usernameline := scanner.Text()
			username, err := util.FormatUsername(usernameline)
			if err != nil {
				keum.logger.Log.Debugf("[!] %q - %v", usernameline, err.Error())
				continue
			}
			time.Sleep(time.Duration(keum.delay) * time.Millisecond)
			usersChan <- username
		}
	}
	close(usersChan)
	wg.Wait()

	finalCount := atomic.LoadInt32(&keum.counter)
	finalSuccess := atomic.LoadInt32(&keum.successes)
	keum.logger.Log.Infof("Done! Tested %d usernames (%d valid) in %.3f seconds", finalCount, finalSuccess, time.Since(start).Seconds())

	if err := scanner.Err(); err != nil {
		keum.logger.Log.Error(err.Error())
	}
}

func (keum *KerberosEnumUsersModule) makeKerberosEnumUsersWorker(ctx context.Context, usernames <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case username, ok := <-usernames:
			if !ok {
				return
			}
			keum.doOneUserEnum(ctx, username)
		}
	}
}

func (keum *KerberosEnumUsersModule) doOneUserEnum(ctx context.Context, username string) {
	atomic.AddInt32(&keum.counter, 1)
	usernamefull := fmt.Sprintf("%v@%v", username, keum.domain)
	valid, rb, err := keum.kSession.TestUsername(username)
	if valid {
		atomic.AddInt32(&keum.successes, 1)
		if rb != nil {
			keum.logger.Log.Infof("[!] VALID USERNAME WITH DONT REQ PREAUTH:\t %s", usernamefull)
		} else {
			if err != nil {
				keum.logger.Log.Infof("[-] VALID USERNAME WITH ERROR:\t %s\t (%s)", username, err)
			} else {
				keum.logger.Log.Noticef("[+] VALID USERNAME BUT WITH REQUIRED PRE AUTH:\t %s", usernamefull)
			}
		}
	} else if err != nil {
		// This is to determine if the error is "okay" or if we should abort everything
		ok, errorString := keum.kSession.HandleKerbError(err)
		if !ok {
			keum.logger.Log.Errorf("[!] %v - %v", usernamefull, errorString)
			keum.cancel()
		} else {
			keum.logger.Log.Debugf("[!] %v - %v", usernamefull, errorString)
		}
	} else {
		keum.logger.Log.Debug("[!] Unknown behavior - %v", usernamefull)
	}
}
