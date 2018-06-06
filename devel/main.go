package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/axamon/sauron2/sms"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

//Reperibile è la variabile con i dati personali dei reperibili
type Reperibile struct {
	id           int
	Nome         string
	Cognome      string
	Cellulare    string
	Assegnazioni Assegnazione
}

//Assegnazione è la variabile con i dati relativi alla ruota di reperibilità
type Assegnazione struct {
	Piattaforma  string
	Giorno       string
	Gruppo       string
	ReperibileID uint
}

var t = time.Now()

//limite delle 7 fino alle 7 del mattino seguente il reperibile che viene visualizzato è quello del giorno prima
var limite7 = time.Date(t.Year(), t.Month(), t.Day(), 7, 0, 0, 0, t.Location())

var ieri = time.Now().Add(-24 * time.Hour).Format("20060102")
var oggi = time.Now().Format("20060102")
var domani = time.Now().Add(24 * time.Hour).Format("20060102")

var filecsv = flag.String("f", "reperibilita.csv", "Percorso del file csv per la reperibilità")
var piattaforma = flag.String("p", "CDN", "La piattaforma di cui desideri ricavare il reperibile")

var contatti []Reperibile

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

const (
	dbfile           = "reperibili.db"
	createreperibile = `
	CREATE TABLE IF NOT EXISTS reperibile (
		id	integer PRIMARY KEY AUTOINCREMENT,
		nome	varchar ( 255 ),
		cognome	varchar ( 255 ),
		cellulare	varchar ( 255 )
	);`

	createassegnazione = `
	CREATE TABLE IF NOT EXISTS assegnazione (
		id	integer PRIMARY KEY AUTOINCREMENT,
		created_at	datetime,
		updated_at	datetime,
		deleted_at	datetime,
		piattaforma	varchar ( 255 ),
		giorno	varchar ( 255 ),
		gruppo	varchar ( 255 ),
		reperibile_id	integer
	);`
)

var db *sql.DB

//InitDB inzializza il database e restituisce la risorsa
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	creadb1, err := db.Prepare(createreperibile)
	checkErr(err)
	_, errcreadb1 := creadb1.Exec()
	checkErr(errcreadb1)
	creadb2, err := db.Prepare(createassegnazione)
	checkErr(err)
	_, errcreadb2 := creadb2.Exec()
	checkErr(errcreadb2)
	addreperibile, err := db.Prepare("INSERT INTO reperibile (id,nome, cognome, cellulare) VALUES (?,?, ?,?)")
	checkErr(err)
	addreperibile.Exec("1", "Alberto", "Bregliano", "+393357291533")
	addreperibile.Exec("2", "Antonio", "Gasponi", "+393357291533")
	return db
}

func main() {
	db := InitDB(dbfile)
	defer db.Close()
	id, err := idRep("Bregliano")
	fmt.Println(id)
	id, err = idRep("Gasponi")
	fmt.Println(id)
	_, id, err = isRepSet("20180606")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(id)
	setRep("20180606", "Bregliano")

}

//addRep Aggiunge un reperibile al DB
func addRep(nome, cognome, cellulare string) (ok bool, err error) {
	if ok := sms.Verificacellulare(cellulare); ok != true {
		return false, fmt.Errorf("Cellulare inserito non nel formato +39(10)cifre")
	}
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Problema ad aprire il DB %s", err.Error())
	}
	defer db.Close()
	verificaprimachenonesistagia, err := db.Prepare("select count(*) from reperibile where nome = ? and cognome = ? and cellulare = ?")
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Problema a preparare la query %s", err.Error())
	}
	addreperibile, err := db.Prepare("INSERT INTO reperibile (nome, cognome, cellulare) VALUES (?, ?,?)")
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Problema a preparare la query %s", err.Error())
	}
	var exist bool
	row := verificaprimachenonesistagia.QueryRow(nome, cognome, cellulare)
	errrow := row.Scan(&exist)
	if errrow != sql.ErrNoRows {
		return false, fmt.Errorf("Impossibile inserire reperibile %s", err.Error())
	}
	_, erraddrep := addreperibile.Exec(nome, cognome, cellulare)
	if erraddrep != nil {
		return false, fmt.Errorf("Impossibile inserire reperibile %s", err.Error())
	}
	return true, nil
}

//setRep assegna un reperibile al giorno
func setRep(giorno, cognome string) (ok bool, err error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Problema ad aprire il DB %s", err.Error())
	}
	defer db.Close()
	idrep, err := idRep(cognome)
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	settaRep, err := db.Prepare("insert into assegnazione (giorno, reperibile_id) values(?,?)")
	if err != nil {
		//fmt.Println(err.Error())
		return false, fmt.Errorf("Problema a preparare la query %s", err.Error())
	}
	_, err = settaRep.Exec(giorno, idrep)
	if err != nil {
		return false, fmt.Errorf("Problema a settare il reperibile %s", err.Error())
	}
	return true, nil

}

//isRepSet informa se un Reperibile è stato impostato per il giorno e qual' è il suo id
func isRepSet(giorno string) (ok bool, reperibileID int, err error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		//fmt.Println(err.Error())
		return false, -1, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	defer db.Close()
	cercagiorno, err := db.Prepare("select reperibile_id from assegnazione where giorno = ?")
	if err != nil {
		return false, -1, fmt.Errorf("errore: %v", err.Error())
	}
	row := cercagiorno.QueryRow(giorno)
	err = row.Scan(&reperibileID)
	if err != nil {
		return false, -1, fmt.Errorf("errore: %v", err.Error())
	}
	return true, reperibileID, nil
}

//infoRep restituisce l'ID del reperibile su DB
func infoRep(idrep int) (info Reperibile, err error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return Reperibile{}, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	defer db.Close()
	retrieveinfo, err := db.Prepare("select nome, cognome, cellulare from reperibile where id = ? limit 1")
	if err != nil {
		//fmt.Println(err.Error())
		return Reperibile{}, fmt.Errorf("Problema con la preparazione della query %s", err.Error())
	}
	row := retrieveinfo.QueryRow(info.Nome, info.Cognome, info.Cellulare)
	err = row.Scan(&info)
	if err != nil {
		//fmt.Println(err.Error())
		return Reperibile{}, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	//fmt.Println(id) //debug
	return info, nil

}

//idRep restituisce l'ID del reperibile su DB
func idRep(cognome string) (id int, err error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return -1, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	defer db.Close()
	retrieveid, err := db.Prepare("select id from reperibile where cognome = ? limit 1")
	if err != nil {
		//fmt.Println(err.Error())
		return -1, fmt.Errorf("Problema con la preparazione della query %s", err.Error())
	}
	row := retrieveid.QueryRow(cognome)
	err = row.Scan(&id)
	if err != nil {
		//fmt.Println(err.Error())
		return -1, fmt.Errorf("Id reperibile non trovato %s", err.Error())
	}
	//fmt.Println(id) //debug
	return id, nil

}
