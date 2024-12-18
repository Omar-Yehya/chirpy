package main 


import ("net/http"
        "log"
        "fmt"
        "sync/atomic"
        "encoding/json"
        "strings"
)


type apiConfig struct{
    fileserverHits atomic.Int32
    
}

type chirpy struct{
    Body string `json:"body"`
}

func main(){


apiCfg:=&apiConfig{}
    

serveMux:=http.NewServeMux()

Filehandler:= http.StripPrefix("/app", http.FileServer(http.Dir(".")))
serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(Filehandler))
serveMux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
serveMux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
serveMux.HandleFunc("/api/validate_chirp/v1", validateChirpHandler)
serveMux.HandleFunc("/api/validate_chirp/v2", validChirp)



serveMux.HandleFunc("GET /api/healthz", healthHandler)


server:=&http.Server{

    Handler: serveMux,
    Addr: ":8080",

    }
log.Fatal(server.ListenAndServe())




}

func healthHandler(w http.ResponseWriter, r * http.Request){
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    cfg.fileserverHits.Add(1)
    next.ServeHTTP(w,r)

    })
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request){

    hitsCount:= cfg.fileserverHits.Load()
    FormattedHTML:=fmt.Sprintf(`
    
        <html>
            <body>
                <h1>Welcome, Chirpy Admin</h1>
                <p>Chirpy has been visited %d times!</p>
            </body>
        </html>
    
    
    `,hitsCount)
    
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(FormattedHTML))


}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request){
 cfg.fileserverHits.Store(0)

}

// checks valid and functional chirpy tweet 
func validateChirpHandler(w http.ResponseWriter, r *http.Request){
 
    decoder:=json.NewDecoder(r.Body)

    params:=chirpy{}
    err:=decoder.Decode(&params)
    if err != nil{
        log.Printf("Error Decoding: %s", err)
        w.WriteHeader(400)
        return

    }


    if  len(params.Body)> 140{
        response:= map[string] string{"error": "Chirp is too long"}
        dat, err := json.Marshal(response)
        if err != nil{
            log.Printf("Error marshalling JSON %s", err)
            w.WriteHeader(500)
            return 
        }
        
    
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(400)
        w.Write(dat)
    }else{
        response:=map[string] bool{"valid": true}
        dat,err:=json.Marshal(response)
        if err != nil{
            log.Printf("Error marshalling JSON %s", err)
            w.WriteHeader(500)
            return 
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write(dat)

    }

}

// Filters bad words 
func validWord(input string) string {

    bad_words:=[]string{"kerfuffle", "sharbert", "fornax"}


    words:=strings.Split(strings.ToLower(input)," ")

    for i,word:= range words{
        for _,badWord:= range bad_words{
            if word ==badWord{
                words[i]="****"
            }
        }
        
    }

    Cencored_text := strings.Join(words, " ")
    return Cencored_text
}

func validChirp(w http.ResponseWriter, r *http.Request){
    var Chirpdata chirpy
    decoder:=json.NewDecoder(r.Body)
    if err:= decoder.Decode(&Chirpdata); err!=nil{
        http.Error(w, "Bad request", http.StatusBadRequest )
        return 
    }

    cleandBody:=validWord(Chirpdata.Body)
    response:=map[string]string{"cleaned_body": cleandBody}

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)





}