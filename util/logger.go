package util

import (
	"os"

	"github.com/op/go-logging"
)

type Logger struct {
	Log *logging.Logger
}

func ASREPRoastingLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("asreproasting")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func KerberosEnumUsersLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("kerberos_enumusers")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func LdapEnumLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("ldap_enum")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func SmbEnumLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("smb_enum")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func KerberoastingLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("kerberoasting")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func KerberosSprayLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("kerberos_password_spraying")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func SmbSprayLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("smb_password_spray")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}

func PayloadGeneratorLogger(verbose bool, logFileName string) Logger {
	log := logging.MustGetLogger("payload_generator")
	format := logging.MustStringFormatter(
		`%{color}%{time:2006/01/02 15:04:05} >  %{message}%{color:reset}`,
	)
	formatNoColor := logging.MustStringFormatter(
		`%{time:2006/01/02 15:04:05} >  %{message}`,
	)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	if logFileName != "" {
		outputFile, err := os.Create(logFileName)
		if err != nil {
			panic(err)
		}
		fileBackend := logging.NewLogBackend(outputFile, "", 0)
		fileFormatter := logging.NewBackendFormatter(fileBackend, formatNoColor)
		logging.SetBackend(backendFormatter, fileFormatter)
	} else {
		logging.SetBackend(backendFormatter)
	}

	if verbose {
		logging.SetLevel(logging.DEBUG, "")
	} else {
		logging.SetLevel(logging.INFO, "")
	}
	return Logger{log}
}
