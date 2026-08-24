package config

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	ColorOff         bool
	Strict           bool
	SuppressWarnings bool
	PreserveWatFile  bool
	NoWASM           bool
	EnableTraces     bool
	WatFile          string
	OutFile          string
	Flags            *flag.FlagSet
}

func LoadAndValidateConfig(args []string, stderr io.Writer) (*Config, error) {
	conf := &Config{}
	fs := flag.NewFlagSet("the", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.BoolVar(&conf.ColorOff, "no-color", false, "Disable color output")
	fs.BoolVar(&conf.SuppressWarnings, "suppress-warnings", false, "Disable reporting of warnings")
	fs.BoolVar(&conf.Strict, "strict", false, "Any warnings will cause compilation to fail")
	fs.BoolVar(&conf.PreserveWatFile, "preserve-wat-file", false, "Writes generated WAT file to disk")
	fs.StringVar(&conf.WatFile, "wat", "", "Path to the generated wat file (if preserved)")
	fs.StringVar(&conf.OutFile, "o", "", "Path to the generated wasm executable")
	fs.BoolVar(&conf.NoWASM, "nowasm", false, "Produce wasm but don't write it to a file")
	fs.BoolVar(&conf.EnableTraces, "enable-traces", false, "Show backtraces on runtime errors (only available in run mode)")
	fs.Usage = FlagUsageMessage(fs)
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	conf.Flags = fs
	err = conf.Validate()
	return conf, err
}

func (conf *Config) Validate() error {
	if conf.Strict && conf.SuppressWarnings {
		return fmt.Errorf("Cannot use strict and suppress-warnings flags together")
	}
	if conf.NoWASM && conf.OutFile != "" {
		return fmt.Errorf("Cannot use nowasm and o flags together")
	}
	if conf.WatFile != "" {
		conf.WatFile = filepath.Clean(conf.WatFile)
		if !strings.HasSuffix(conf.WatFile, ".wat") {
			return fmt.Errorf("wat file must have '.wat' extension")
		}
	}
	if conf.OutFile != "" {
		conf.OutFile = filepath.Clean(conf.OutFile)
		if !strings.HasSuffix(conf.OutFile, ".wasm") {
			return fmt.Errorf("Output file must have '.wasm' extension")
		}
	}
	return nil
}

func FlagUsageMessage(fs *flag.FlagSet) func() {
	return func() {
		output := fs.Output()
		fmt.Fprintf(output, "Usage: %s [version | help]\n", os.Args[0])
		fmt.Fprintf(output, "Usage: %s [options] [run | build] [file]\n", os.Args[0])
		fmt.Fprintln(output, "options:")
		fs.PrintDefaults()
	}
}

type EnvConfig struct {
	Stdout        io.Writer
	Stderr        io.Writer
	LogBuffer     bytes.Buffer
	DevMode_Lexer bool
	DevMode_AST   bool // after parser
	// after semantic
	DevMode_ScopeTree    bool
	DevMode_AnnotatedAST bool
	DevMode_irgen        bool
	DevMode_codegen      bool
	Debug                bool
}

func LoadEnvConfig(stdout, stderr io.Writer) *EnvConfig {
	conf := &EnvConfig{}
	conf.Stdout = stdout
	conf.Stderr = stderr
	log.SetFlags(log.Lshortfile)
	log.SetOutput(&conf.LogBuffer)
	vars := map[string]*bool{
		"THE_DEV_LEXER":     &conf.DevMode_Lexer,
		"THE_DEV_PARSER":    &conf.DevMode_AST,
		"THE_DEV_SCOPES":    &conf.DevMode_ScopeTree,
		"THE_DEV_ANNOTATED": &conf.DevMode_AnnotatedAST,
		"THE_DEV_IRGEN":     &conf.DevMode_irgen,
		"THE_DEV_CODEGEN":   &conf.DevMode_codegen,
		"THE_DEV_DEBUG":     &conf.Debug,
	}
	for key, flag := range vars {
		value, err := fetchEnvironmentVariableValue(key)
		if err != nil {
			*flag = false
		} else {
			*flag = value
		}
	}
	return conf
}

func fetchEnvironmentVariableValue(key string) (bool, error) {
	envVar := os.Getenv(key)
	return strconv.ParseBool(envVar)
}

func (conf EnvConfig) PrintDebugLogs() {
	if conf.Debug && conf.LogBuffer.Len() != 0 {
		fmt.Fprintln(conf.Stdout, conf.LogBuffer.String())
	}
}
