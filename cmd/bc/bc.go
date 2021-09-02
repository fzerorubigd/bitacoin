package main

import (
	"flag"
	"github.com/fzerorubigd/bitacoin"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/hash/{hash}", bitacoin.HashDetail)
	r.HandleFunc("/", bitacoin.Index)
	r.HandleFunc("/transfer", bitacoin.Transfer)
	r.HandleFunc("/balance/{owner}", bitacoin.Balance)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("template/static"))))

	http.Handle("/", r)

	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		return
	}

	//flag.Usage = usage
	//flag.Parse()

	//Run file base
	//runByFile()

	//Run db format
	//runBoltDB()
}

func runByFile() {
	var store string
	flag.StringVar(&store, "store", os.Getenv("BC_STORE"), "The store to use")
	s := bitacoin.NewFolderStore(store)

	if err := dispatch(s, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}

func runBoltDB() {

	d := bitacoin.NewDBStore()
	if err := dispatch(d, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
