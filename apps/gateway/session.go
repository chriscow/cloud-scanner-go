package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/chriscow/cloud-scanner-go/geom"
	"github.com/chriscow/cloud-scanner-go/scan"

	"github.com/go-chi/render"
	"github.com/nsqio/go-nsq"
)

// SessionPayload ...
type SessionPayload struct {
	*scan.Session
}

// Bind on SessionPayload allows post-processing after unmarshalling
func (s *SessionPayload) Bind(r *http.Request) error {
	if s.Session == nil {
		return errors.New("missing required Session field")
	}

	if s.Session.ID == 0 {
		s.Session.ID = time.Now().UnixNano()
	}

	return nil
}

// Render on SessionPayload allows pre-processing before a response is marshalled
// and sent across the wire
func (s *SessionPayload) Render(w http.ResponseWriter, r *http.Request) error {
	// s.Elapsed = ... for example
	return nil
}

func getDefaultSession(w http.ResponseWriter, r *http.Request) {
	session := &scan.Session{
		ZLine: geom.ZLine{
			Limit: 100,
			Zeros: []geom.Zeros{
				geom.Zeros{
					ZeroType: geom.Primes,
					Scalar:   1,
				},
			},
		},
		Radius:        1,
		DistanceLimit: 1,
		BucketCount:   3600,
		ScansReq:      5000,
		MinScore:      .3, // 30% of zeros were hit. Cannot be zero
	}

	payload := SessionPayload{Session: session}

	render.Render(w, r, &payload)
}

func startSession(w http.ResponseWriter, r *http.Request) {
	payload := &SessionPayload{}
	if err := render.Bind(r, payload); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		render.Render(w, r, ErrServerError("NewProducer", err))
	}
	defer producer.Stop()

	body, err := json.Marshal(*payload.Session)
	if err != nil {
		render.Render(w, r, ErrServerError("Marshal", err))
	}

	err = producer.Publish(scan.SessionTopic, body)
	if err != nil {
		render.Render(w, r, ErrServerError("Publish", err))
		return
	}

	render.Status(r, http.StatusCreated)
	render.Render(w, r, payload)
}

func getSession(w http.ResponseWriter, r *http.Request) {
}

// scanLatticeCmd generates scan-radius sessions and publishes them to the
// channel returned.  Each session contains a different origin such that all the
// scan sessions will completely cover the lattice.
/*
func scanLatticeCmd(ctx *cli.Context) error {
	if ctx.NArg() < 2 {
		return errors.New("Expected lattice and one or more zeros")
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Instantiate a producer.
	topic := "scan-session"
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	// We create one session, thus only loading the lattice and zeros once
	// then just modify its ID and zline origin in the loop below
	cctx, _ := context.WithCancel(context.Background()) // actually unused
	s, err := sessionFromCLI(cctx, ctx)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	origins := s.Lattice.Partition(s.Radius)
	elapsed := time.Since(start)
	log.Println("Lattice partitioned in", elapsed.Seconds())

	wg := &sync.WaitGroup{}
	wg.Add(len(origins))

	start = time.Now()
	for id, origin := range origins {
		select {
		case <-sigChan:
			msg := fmt.Sprint("\nCanceled by user")
			log.Fatal(msg)
		default:
			go func() {

				s.ID = start.UnixNano() + int64(id)
				s.ZLine.Origin = origin

				body, err := json.Marshal(s)
				if err != nil {
					log.Fatal(err)
				}

				err = producer.Publish(topic, body)
				if err != nil {
					log.Fatal(err)
				}

				wg.Done()
			}()
		}
	}

	wg.Wait()
	elapsed = time.Since(start)
	log.Println("Published", len(origins), "sessions in", elapsed.Seconds())

	return nil
}
*/
