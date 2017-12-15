package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Comcast/sheens/cmd/mservice/storage/bolt"
	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"
	"github.com/Comcast/sheens/interpreters/goja"
)

var (
	DefaultSpecDir = "../../specs"
	DefaultLibDir  = "../.."
)

func makeDemoService(ctx context.Context, routed chan interface{}, specDir, libDir string) (*Service, error) {

	if specDir == "" {
		specDir = DefaultSpecDir
	}

	log.Printf(`specDir: "%s"`, specDir)

	if libDir == "" {
		libDir = DefaultLibDir
	}

	log.Printf(`libDir: "%s"`, specDir)

	bs, err := bolt.NewStorage("storage.db")
	if err != nil {
		return nil, err
	}
	// bs.Debug = true
	if err = bs.Open(); err != nil {
		return nil, err
	}

	gi := goja.NewInterpreter()
	gi.LibraryProvider = goja.MakeFileLibraryProvider(libDir)

	interpreters := map[string]core.Interpreter{
		"goja": gi,
	}

	s, err := NewService()
	if err != nil {
		return nil, err
	}

	// An a crude example, here's an in-process HTTP request
	// service.  When a machine emits a message that includes
	// "to":{"mid":"HTTP"}, then this thing will process that
	// message.  That processing involves trying to parse the
	// message as an HTTPRequest.  Then making the request.  Then
	// sending the response back to the crew that made the
	// request.
	//
	// Obviously this example could be extended and refined
	// considerably.  For example, the requester could optionally
	// provide a specific machine id to receive the response.
	httpRequestListener := HTTPRequestListener{
		// This function will take an HTTPResponse as input
		// (as the message) and send it to the right crew.
		Emitter: func(ctx context.Context, c *crew.Crew, message interface{}) {
			log.Printf("inbound %s", JS(message))
			go func() {
				walkeds, err := s.Process(ctx, c.Id, message, nil)
				if err != nil {
					log.Printf("HTTP response process error %v", err)
				} else {
					log.Printf("HTTP response processed %s", JS(walkeds))
				}
			}()
		},
	}

	// router will direct a message that include
	// "to":{"mid":"?mid"} to a machine with id "?mid". Otherwise,
	// messages go to all Machines in the Crew.
	//
	// This function also sends messages to the routed channel (if
	// it's not nil), so we can easily watch what is routed.
	router := func(ctx context.Context, c *crew.Crew, message interface{}) ([]string, bool, error) {
		log.Printf("routing %s", JS(message))

		if routed != nil {
			select {
			case routed <- message:
			default:
				log.Printf("routed channel is clogged")
			}
		}

		bss, err := core.Match(nil, Dwimjs(`{"to":{"mid":"?mid"}}`), message, core.NewBindings())
		if err != nil {
			return nil, false, err
		}
		if 0 < len(bss) {
			x, _ := bss[0]["?mid"]
			if mid, is := x.(string); is {
				log.Printf("routing direct to %s", mid)
				switch mid {
				case "HTTP":
					httpRequestListener.Handle(ctx, c, message)
					return nil, false, nil
				default:
					return []string{mid}, false, nil
				}
			}
			return nil, false, fmt.Errorf("bad mid %T: %#v", x, x)
		}

		return nil, true, nil
	}

	s.Interpreters = interpreters
	s.Router = router
	s.Storage = bs

	// Our specs will just live on the file system.
	specProvider, err := NewFileSystemSpecProvider(ctx, s, specDir)
	if err != nil {
		return nil, err
	}
	s.SpecProvider = specProvider.Find

	if err = specProvider.ReadSpecs(ctx); err != nil {
		return nil, err
	}

	return s, nil
}
