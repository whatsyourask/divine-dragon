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

var (
	domain       string
	dc           string
	verbose      bool
	safemode     bool
	downgrade    bool
	usernamelist string

	kSession transport.KerberosSession
	kOptions transport.KerberosSessionOptions

	ctx, cancel = context.WithCancel(context.Background())
	threads     int
	delay       int
	counter     int32
	successes   int32

	logger util.Logger
)

func setupSession() bool {
	kOptions = transport.KerberosSessionOptions{
		Domain:           domain,
		DomainController: dc,
		Verbose:          verbose,
		SafeMode:         safemode,
		Downgrade:        downgrade,
	}
	k, err := transport.NewKerberosSession(kOptions)
	if err != nil {
		logger.Log.Error(err)
		return false
	}
	kSession = k

	logger.Log.Info("Using KDC(s):")
	for _, v := range kSession.Kdcs {
		logger.Log.Infof("\t%s\n", v)
	}

	return true
}

func SetupKerberosEnumUsersModule(domain_opt string, dc_opt string, verbose_opt bool, safemode_opt bool,
	downgrade_opt bool, usernamelist_opt string, logFileName string, threads_opt int, delay_opt int) bool {
	domain = domain_opt
	dc = dc_opt
	verbose = verbose_opt
	safemode = safemode_opt
	downgrade = downgrade_opt
	usernamelist = usernamelist_opt

	delay = delay_opt
	if delay != 0 {
		threads = 1
		logger.Log.Infof("Delay set. Using single thread and delaying %dms between attempts\n", delay)
	} else {
		threads = threads_opt
	}
	logger = util.KerberosEnumUsersLogger(verbose, logFileName)
	status := setupSession()
	return status
}

func KerberosEnumUsers() {
	usersChan := make(chan string, threads)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(threads)

	var scanner *bufio.Scanner
	if usernamelist != "-" {
		file, err := os.Open(usernamelist)
		if err != nil {
			// logger.Log.Error(err.Error())
			return
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	for i := 0; i < threads; i++ {
		go makeKerberosEnumUsersWorker(ctx, usersChan, &wg)
	}

	start := time.Now()

Scan:
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			break Scan
		default:
			usernameline := scanner.Text()
			username, err := util.FormatUsername(usernameline)
			if err != nil {
				logger.Log.Debugf("[!] %q - %v", usernameline, err.Error())
				continue
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
			usersChan <- username
		}
	}
	close(usersChan)
	wg.Wait()

	finalCount := atomic.LoadInt32(&counter)
	finalSuccess := atomic.LoadInt32(&successes)
	logger.Log.Infof("Done! Tested %d usernames (%d valid) in %.3f seconds", finalCount, finalSuccess, time.Since(start).Seconds())

	if err := scanner.Err(); err != nil {
		logger.Log.Error(err.Error())
	}
}

func makeKerberosEnumUsersWorker(ctx context.Context, usernames <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case username, ok := <-usernames:
			if !ok {
				return
			}
			doOneKerberosEnumUser(ctx, username)
		}
	}
}

func doOneKerberosEnumUser(ctx context.Context, username string) {
	atomic.AddInt32(&counter, 1)
	usernamefull := fmt.Sprintf("%v@%v", username, domain)
	valid, rb, err := kSession.TestUsername(username)
	if valid {
		atomic.AddInt32(&successes, 1)
		if rb != nil {
			logger.Log.Infof("[!] VALID USERNAME WITH DONT REQ PREAUTH:\t %s", usernamefull)
		} else {
			if err != nil {
				logger.Log.Infof("[-] VALID USERNAME WITH ERROR:\t %s\t (%s)", username, err)
			} else {
				logger.Log.Noticef("[+] VALID USERNAME PRE AUTH REQUIRED:\t %s", usernamefull)
			}
		}
	} else if err != nil {
		// This is to determine if the error is "okay" or if we should abort everything
		ok, errorString := kSession.HandleKerbError(err)
		if !ok {
			logger.Log.Errorf("[!] %v - %v", usernamefull, errorString)
			cancel()
		} else {
			logger.Log.Debugf("[!] %v - %v", usernamefull, errorString)
		}
	} else {
		logger.Log.Debug("[!] Unknown behavior - %v", usernamefull)
	}
}
