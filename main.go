// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/minio/pkg/env"
)

var (
	logFile   string
	address   string
	authToken = env.Get("WEBHOOK_AUTH_TOKEN", "")
)

func main() {
	flag.StringVar(&logFile, "log-file", "", "path to the file where webhook will log incoming events")
	flag.StringVar(&address, "address", ":8080", "bind to a specific ADDRESS:PORT, ADDRESS can be an IP or hostname")

	flag.Parse()

	if logFile == "" {
		log.Fatalln("--log-file must be specified")
	}

	l, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		log.Fatal(err)
	}

	var mu sync.Mutex

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for _ = range sigs {
			mu.Lock()
			l.Sync()  // flush to disk any temporary buffers.
			l.Close() // then close the file, before rotation.
			l, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
			if err != nil {
				log.Fatal(err)
			}
			mu.Unlock()
		}
	}()

	err = http.ListenAndServe(address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authToken != "" {
			if authToken != r.Header.Get("Authorization") {
				http.Error(w, "authorization header missing", http.StatusBadRequest)
				return
			}
		}
		switch r.Method {
		case http.MethodPost:
			mu.Lock()
			_, err := io.Copy(l, r.Body)
			if err != nil {
				mu.Unlock()
				return
			}
			l.WriteString("\n")
			mu.Unlock()
		default:
		}
	}))
	if err != nil {
		log.Fatal(err)
	}
}
