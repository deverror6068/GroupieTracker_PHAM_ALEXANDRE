package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

var home = template.Must(template.ParseFiles("ressources/index.html")) //page principale

var index = template.Must(template.ParseFiles("ressources/filmid.html")) //page de recherche par id

var result = template.Must(template.ParseFiles("ressources/result.html")) // page de resultat de la recherche par id

var result2 = template.Must(template.ParseFiles("ressources/results2.html")) // page de resultat du titre

var title = template.Must(template.ParseFiles("ressources/title.html")) //  page de recherche par titre

type Apifield struct {
	Title        string `json:"Title"`      //titre du film
	ImbRated     string `json:"imdbRating"` //note du film
	Query        string // données entrées par l'utilisateur
	Response     string `json:"Response"` //reponse de la requete
	Error        string `json:"Error"`    //erreur et nom  de l'erreur
	Request      string
	Plot         string `json:"Plot"`    // synopsis
	Runtime      string `json:"Runtime"` //durée du film
	Poster       string `json:"Poster"`  //affiche du film
	Genre        string `json:"Genre"`
	Noimage      bool   // si il n'y  pas  d'image disponible
	Niadress     string // 	adresse url de l'image de remplacement
	Validrequest bool   // la requete est elle valide
	Awards       string `json:"Awards"`
	Result2title string
}

func main() {

	log.Println("Debut du Programme")

	styleServer := http.FileServer(http.Dir("css"))              // indication du fichier contenant le dossier
	http.Handle("/css/", http.StripPrefix("/css/", styleServer)) // gestion du css par le serveur

	http.HandleFunc("/", HttpHandlertitle) // route du serveur + fonction associée au serveur
	http.HandleFunc("/home", HttpHandlehome)
	http.HandleFunc("/resultsid", HttpHandlerresult)
	http.HandleFunc("/resultstitle", HttpHandlerresulttitle)
	http.HandleFunc("/title", HttpHandlertitlepage)
	http.ListenAndServe(":80", nil) //lancement du serveur sur le port 80

}

func HttpHandlertitle(w http.ResponseWriter, r *http.Request) {

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666) //ouverture-cration du fichier qui va stocker la valeur de l'utilisateur
	if err != nil {
		fmt.Println("erreur lors de l'ouverture du fichier log ")
		err = exec.Command("cmd.exe", "/c", "exitcode15.vbs").Run() //Affichage de l'erreur à l'utilisateur (plus pratique)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(15) // code d'erreur 15
	}

	log.SetOutput(file)

	url1 := "https://www.omdbapi.com/?i=tt" // premiere partie de la requete
	url2 := "&apikey=e78ca7eb"              //deuxième partie de la requete
	query := r.FormValue("w")               // recupération de la valeur entrée par l'utilisateur qui à le nom/name w dans le fichier html
	query1 := query
	io.Copy(file, strings.NewReader(query1)) //copie de la valeur l'id  donnée par l'utilisateur

	url1 += query
	url1 += url2
	defer file.Close() //fermeture du fichiers pour pouvoir le supprimer dans la template results
	var url string
	url = url1

	timeClient := http.Client{ //definition du delai de time out du serveur
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil) //envoi de requete
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "spacecount-total")

	res, getErr := timeClient.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body) //lecture du body
	if readErr != nil {
		fmt.Println(readErr)
	}

	var temp1 Apifield
	fmt.Sprintln(temp1.Poster)
	total := json.Unmarshal(body, &temp1)

	if total != nil {
		fmt.Println(temp1.ImbRated)
	}

	if temp1.Poster == "N/A" { //si il n'y a pas d'image pour le film
		temp1.Noimage = true
		temp1.Niadress = "https://www.radiobeton.com/www/wp-content/uploads/2017/01/arton17969.jpg" //affichage d'une image de remplacement

	}

	if temp1.Response == "True" { //verififcation  de validité de la requette  (le html affichera "plus de details" si true)
		temp1.Validrequest = true
	}

	data := Apifield{
		Title:        fmt.Sprintf(temp1.Title), //affichage du titre avec sprintf
		ImbRated:     fmt.Sprintf(temp1.ImbRated),
		Query:        fmt.Sprintf(query),
		Response:     fmt.Sprintf(temp1.Response),
		Error:        fmt.Sprintf(temp1.Error),
		Plot:         fmt.Sprintf(temp1.Plot),
		Genre:        fmt.Sprintf(temp1.Genre),
		Poster:       fmt.Sprintf(temp1.Poster),
		Noimage:      temp1.Noimage,
		Niadress:     fmt.Sprintf(temp1.Niadress),
		Validrequest: temp1.Validrequest,
	}

	index.Execute(w, data)
}

func HttpHandlerresult(w http.ResponseWriter, r *http.Request) { //fonction de la page resultat (id)

	read, err := ioutil.ReadFile("logs.txt") //lecture du contenu

	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier", err)
		err = exec.Command("cmd.exe", "/c", "exitcode16.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(16)
	}

	url1 := "https://www.omdbapi.com/?i=tt"
	url2 := "&apikey=e78ca7eb"

	request := string(read) //conversion  de  read en string

	void := ""            // la variable contiendra la dernière requete
	if len(request) > 7 { // si le taille de la requete de l'id dépasse 7 (si il y a plusieurs requete de l'utilisateur) le code ci-dessous permet de selectionner la requete la plus recente
		un := 0
		sept := 0
		septinrequest := len(request) / 7 //determine le nombre de requetes

		for i := 0; i < len(request); i++ { // parcours de la requete
			un = un + 1
			if un == 7 { //si l'on passe à la requete suivante
				sept += 1
				un = 0
				print("test", un, sept)
			}
			if sept == septinrequest-1 && i >= (septinrequest-1)*7 || sept == septinrequest { //si l'on passe à la dernière requete
				fmt.Printf("%x ", request[i])
				void += string(request[i]) //on concatene

			}

		}

		request = void
	}

	url1 += request
	url1 += url2

	err1 := os.Remove("logs.txt") // on supprime le fichier  pour actualiser les données rentrées

	if err1 != nil {
		fmt.Println("erreur lors de la manipulation du fichier : ", err1) //gestion d'erreur
		err = exec.Command("cmd.exe", "/c", "exitcode17.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(17)
	}

	f, err := os.Create("logs.txt")
	if err != nil {
		fmt.Println("erreur lors de la création du fichier logs.txt", err) //gestion d'erreur
		err = exec.Command("cmd.exe", "/c", "exitcode18.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(18)
	}

	defer f.Close() // on ferme le fichier pour eviter les problèmes de type le fichier est ouvert par un autre processus
	var url string
	url = url1

	timeClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "spacecount-total")

	res, getErr := timeClient.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(readErr)
	}
	var temp2 Apifield //variable locale

	total := json.Unmarshal(body, &temp2)

	if total != nil {
		fmt.Println(temp2.ImbRated)
	}

	data := Apifield{ //donnée de la page
		Title:    fmt.Sprintf(temp2.Title),
		ImbRated: fmt.Sprintf(temp2.ImbRated),
		Response: fmt.Sprintf(temp2.Response),
		Error:    fmt.Sprintf(temp2.Error),
		Request:  fmt.Sprintf(temp2.Response, "Response"),
		Plot:     fmt.Sprintf(temp2.Plot),
		Query:    fmt.Sprintf(request),
		Poster:   fmt.Sprintf(temp2.Poster),
		Genre:    fmt.Sprintf(temp2.Genre),
		Runtime:  fmt.Sprintf(temp2.Runtime),
		Awards:   fmt.Sprintf(temp2.Awards),
	}

	result.Execute(w, data)

}

func HttpHandlertitlepage(w http.ResponseWriter, r *http.Request) { // page de recherche par titre
	file, err := os.OpenFile("titlepart1.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("erreur lors de l'ouverture du fichier titlepart1.txt ")
		err = exec.Command("cmd.exe", "/c", "exitcode15-1.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(15) // creation des fichiers qui contiendront les deux élèments de requete de  l'utilisateur
	}

	log.SetOutput(file)

	file1, err := os.OpenFile("titlepart2.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("erreur lors de l'ouverture du fichier titlepart2.txt ")
		err = exec.Command("cmd.exe", "/c", "exitcode15-2.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(15)
	}
	log.SetOutput(file1)

	url1 := "http://www.omdbapi.com/?t="
	urlyear := "&y"
	url2 := "&apikey=e78ca7eb"
	querytitle := r.FormValue("y") // meme fonctionnnenment que sur la page de rechercher par id mais de 2 élèments
	queryyear := r.FormValue("z")
	querytitle1 := querytitle + "/" // ajout d'un / pour séparer les différentes requetes de l'utilisateur dans le fichier
	io.Copy(file, strings.NewReader(querytitle1))
	io.Copy(file1, strings.NewReader(queryyear))

	url1 += querytitle
	url1 += urlyear
	url1 += queryyear
	url1 += url2
	defer file.Close()
	defer file1.Close() //fermeture du fichiers pour pouvoir le supprimer dans la template results
	var url string
	url = url1

	timeClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "spacecount-total")

	res, getErr := timeClient.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}

	// nouvelle requete

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(readErr)
	}

	var temp1 Apifield
	fmt.Sprintln(temp1.Poster)
	total := json.Unmarshal(body, &temp1)

	if total != nil {
		fmt.Println(temp1.ImbRated)
	}
	// si il n'y pas d'image pour le film
	if temp1.Poster == "N/A" {
		temp1.Noimage = true
		temp1.Niadress = "https://www.radiobeton.com/www/wp-content/uploads/2017/01/arton17969.jpg"

	}

	if temp1.Response == "True" {
		temp1.Validrequest = true
	}
	data := Apifield{ //données de la page
		Title:        fmt.Sprintf(temp1.Title),
		ImbRated:     fmt.Sprintf(temp1.ImbRated),
		Query:        fmt.Sprintf(querytitle),
		Response:     fmt.Sprintf(temp1.Response),
		Error:        fmt.Sprintf(temp1.Error),
		Plot:         fmt.Sprintf(temp1.Plot),
		Runtime:      fmt.Sprintf(temp1.Runtime),
		Awards:       fmt.Sprintf(temp1.Awards),
		Genre:        fmt.Sprintf(temp1.Genre),
		Poster:       fmt.Sprintf(temp1.Poster),
		Noimage:      temp1.Noimage,
		Niadress:     fmt.Sprintf(temp1.Niadress),
		Validrequest: temp1.Validrequest,
	}

	title.Execute(w, data)

}

func HttpHandlerresulttitle(w http.ResponseWriter, r *http.Request) {
	read, err := ioutil.ReadFile("titlepart1.txt") //lecture du contenu

	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier", err)
		err = exec.Command("cmd.exe", "/c", "exitcode16-1.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(16)
	}

	read2, err := ioutil.ReadFile("titlepart2.txt") //lecture du contenu et recuperation de la requete utilisateur

	if err != nil {
		fmt.Println("Erreur lors de la lecture du fichier", err)
		err = exec.Command("cmd.exe", "/c", "exitcode16-2.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(16)
	}

	url1 := "http://www.omdbapi.com/?t="
	urlyear := "&y"
	url2 := "&apikey=e78ca7eb"
	querytitle := string(read)
	var slash int = 0    // variable pour compter le nb de /
	var slashcal int = 0 // variable qui permet de savoir ou en est le parcours des requetes
	var tempstring string = ""
	for i := 0; i < len(querytitle); i++ { //lecture du fichier
		if string(querytitle[i]) == "/" { // si il y a le / séparateur

			slash = slash + 1
		}

	}
	for j := 0; j < len(querytitle); j++ {
		if string(querytitle[j]) == "/" {

			slashcal = slashcal + 1
		}
		if slash-1 == slashcal { // si on arrive à la dernière requete (la plus recente)
			tempstring += string(querytitle[j]) // on la recupère

		}
	}
	var tempstring2 = ""

	for g := 0; g < len(tempstring); g++ {
		if string(tempstring[g]) != "/" { // on enlève le / restant
			tempstring2 += string(tempstring[g])
		}

	}
	tempstring = tempstring2 // on donne uniquement la requete utilisateur la plus récente

	queryyear := string(read2)
	querytitle = tempstring
	querytitle1 := querytitle

	url1 += querytitle // concatenation avec l'url
	url1 += urlyear
	url1 += queryyear
	url1 += url2

	err1 := os.Remove("titlepart1.txt") //supprime le fichier pour eviter des temps d'execution trop longs lors le de la lecture du fichier et d'autres effets indésirables

	if err1 != nil {
		fmt.Println("erreur lors de la manipulation du fichier titlepart1.txt: ", err1)
		err = exec.Command("cmd.exe", "/c", "exitcode17-1.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		os.Exit(17)
	}

	f, err := os.Create("titlepart1.txt") // recrée le fichier vierge pour les prochaines requetes
	if err != nil {
		fmt.Println("erreur lors de la création du fichier titlepart1.txt", err)
		err = exec.Command("cmd.exe", "/c", "exitcode18-1.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
		os.Exit(18)
	}

	defer f.Close()

	err2 := os.Remove("titlepart2.txt")

	if err2 != nil {
		fmt.Println("erreur lors de la manipulation du fichier titlepart1.txt : ", err2) //supprime le fichier pour eviter des temps d'execution trop longs lors le de la lecture du fichier et d'autres effets indésirables
		err = exec.Command("cmd.exe", "/c", "exitcode17-2.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		os.Exit(17)
	}

	f1, err := os.Create("titlepart2.txt")
	if err != nil {
		fmt.Println("erreur lors de la création du fichier titlepart2.txt", err) // recrée le fichier vierge pour les prochaines requetes
		err = exec.Command("cmd.exe", "/c", "exitcode18-2.vbs").Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

	}

	defer f1.Close()
	var url string
	url = url1

	timeClient := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "spacecount-total")

	res, getErr := timeClient.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(readErr)
	}
	var temp2 Apifield
	temp2.Result2title = tempstring

	total := json.Unmarshal(body, &temp2)

	if total != nil {
		fmt.Println(temp2.ImbRated)
	}

	data := Apifield{
		Title:        fmt.Sprintf(temp2.Title),
		ImbRated:     fmt.Sprintf(temp2.ImbRated),
		Response:     fmt.Sprintf(temp2.Response),
		Error:        fmt.Sprintf(temp2.Error),
		Request:      fmt.Sprintf(temp2.Response),
		Query:        fmt.Sprintf(querytitle, querytitle1),
		Poster:       fmt.Sprintf(temp2.Poster),
		Plot:         fmt.Sprintf(temp2.Plot),
		Awards:       fmt.Sprintf(temp2.Awards),
		Runtime:      fmt.Sprintf(temp2.Runtime),
		Genre:        fmt.Sprintf(temp2.Genre),
		Result2title: fmt.Sprintf(temp2.Result2title),
	}

	result2.Execute(w, data)

}

func HttpHandlehome(w http.ResponseWriter, r *http.Request) {

	data := Apifield{}

	home.Execute(w, data)
}
