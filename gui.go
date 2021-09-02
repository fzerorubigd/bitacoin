package bitacoin

import (
	"encoding/hex"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Money struct {
	From  string
	To    string
	Money string
}

func tG(name string) *pongo2.Template {
	return pongo2.Must(pongo2.FromFile(fmt.Sprintf("template/%s.html", name)))
}

func Index(w http.ResponseWriter, r *http.Request) {
	store := NewDBStore()
	bc, err := OpenBlockChain(Difficulty, store)
	if err != nil {
		_, err := NewBlockChain([]byte("bita"), Difficulty, store)
		if err != nil {
			fmt.Printf("genesis failed: %w", err)
		}
	}

	blocks := bc.AllBlocks()

	err = tG("main").ExecuteWriter(pongo2.Context{"blocks": blocks}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func HashDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	store := NewDBStore()
	hash, errr := hex.DecodeString(vars["hash"])
	if errr != nil {

	}

	b, err := store.Load(hash)
	if err != nil {
		log.Fatalln(err)
	}

	err = tG("hashDetail").ExecuteWriter(pongo2.Context{"b": b}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := tG("transfer").ExecuteWriter(pongo2.Context{}, w)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if r.Method == "POST" {
		store := NewDBStore()
		from := r.FormValue("from")
		to := r.FormValue("to")
		amount := r.FormValue("amount")
		bc, err := OpenBlockChain(Difficulty, store)
		if err != nil {
			//log.Fatalln("open failed: %w", err)
		}
		i, err := strconv.ParseInt(amount, 10, 32)

		txn, err := NewTransaction(bc, []byte(from), []byte(to), int(i))
		if err != nil {
			store := NewDBStore()
			bc, err2 := OpenBlockChain(Difficulty, store)
			if err2 != nil {
				log.Fatalln(err)
			}

			blocks := bc.AllBlocks()
			err4 := tG("main").ExecuteWriter(pongo2.Context{"blocks": blocks, "error": err.Error()}, w)
			if err4 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}

		reward := NewCoinBaseTxn([]byte("bita"), nil)
		_, err = bc.Add(txn, reward)
		if err != nil {
			log.Fatalln(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func Balance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	store := NewDBStore()
	bc, err := OpenBlockChain(Difficulty, store)
	if err != nil {
		log.Fatalln(err)
	}

	tr, _, acc, err := bc.UnspentTxn([]byte(vars["owner"]))
	if err != nil {
		log.Fatalln(err)
	}

	err = tG("balance").ExecuteWriter(pongo2.Context{"acc": acc, "transaction": tr, "name": vars["owner"]}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
