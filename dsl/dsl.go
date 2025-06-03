package dsl

import (
	"autofat/config"
	"autofat/dsl/parser"
	"autofat/dsl/runtime"
	"autofat/dsl/scanner"
	"autofat/dsl/storage"
	"autofat/dsl/tokens"
	"autofat/logger"
	"log"
	"os"
	"time"
)

func Load(test_file string, config config.Config) (bool, error) {
	preamble, err := os.ReadFile("common")
	if err != nil {
		return false, err
	}

	test_src, err := os.ReadFile("testfiles/" + test_file)
	if err != nil {
		return false, err
	}

	file_contents := string(preamble) + string(test_src)
	start := time.Now()
	_, scanner_stream := scanner.Lex(string(file_contents))

	_exit := false
	word_stream := make([]tokens.Token, 0)
	for !_exit {
		c := <-scanner_stream
		switch c.Symbol {
		case tokens.ItemEOF:
			word_stream = append(word_stream, c)
			_exit = true
		case tokens.ItemError:
			log.Fatal(c)
			return false, err
		default:
			word_stream = append(word_stream, c)
		}
	}

	logger.Log(logger.PERF, "Scanned in ", time.Since(start))

	grammar := tokens.NewGrammar(tokens.NTGoal)
	cfg := parser.CreateCFG()

	start = time.Now()
	parser := parser.CreateLRParser(grammar, cfg, parser.First(cfg, grammar), config.CompileParser)
	logger.Log(logger.PERF, "Created parse tables in ", time.Since(start))

	words := make(chan tokens.Token)
	go func() {
		for i := range word_stream {
			words <- word_stream[i]
		}
	}()
	storage := storage.NewStorage()
	runtime := runtime.New(config)

	generateGlobalVariables(&storage)
	generateGlobalFunctions(runtime, &storage)

	start = time.Now()
	entryPoint, err := parser.Parse(words, cfg, grammar, &storage, runtime)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log(logger.PERF, "Parsed in ", time.Since(start))

	start = time.Now()
	primary := runtime.NewInstance(entryPoint, nil)
	done := make(chan bool)
	kill := make(chan bool)
	go primary.Run(kill, done)
	output := <-done
	logger.Log(logger.PERF, "Program finished in", time.Since(start))

	return output, nil
}
