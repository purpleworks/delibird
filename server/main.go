package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/codegangsta/negroni"
	"github.com/danryan/env"
	"github.com/gorilla/mux"
	"github.com/purpleworks/delibird"
	"github.com/rs/cors"
	"github.com/unrolled/render"
)

type Config struct {
	Environment string `env:"key=ENVIRONMENT default=development"`
	Port        string `env:"key=PORT default=9000"`
	EnableCors  string `env:"key=ENABLE_CORS default=false"`
}

var (
	renderer *render.Render
	config   *Config
)

func init() {
	var option render.Options
	config = &Config{}
	if err := env.Process(config); err != nil {
		fmt.Println(err)
	}
	if config.Environment == "development" {
		option.IndentJSON = true
	}
	renderer = render.New(option)
}

func renderErrorJson(w http.ResponseWriter, err *delibird.ApiError, status int) {
	if status < 100 {
		status = http.StatusBadRequest
	}

	renderer.JSON(w, status, map[string]string{"code": err.Code, "message": err.Message})
}

func App() http.Handler {
	// router
	r := mux.NewRouter()
	r.HandleFunc("/tracking/{code}/{trackingNumber}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		courier, err := delibird.NewCourier(vars["code"])
		if err != nil {
			fmt.Println(err)
			renderErrorJson(w, err, 400)
			return
		}

		trackingNumber := strings.Replace(vars["trackingNumber"], "-", "", -1)
		data, err := courier.Parse(trackingNumber)
		if err != nil {
			renderErrorJson(w, err, 400)
			return
		}

		renderer.JSON(w, http.StatusOK, data)
	})

	// middleware
	n := negroni.Classic()

	// enable CORS
	if config.EnableCors == "true" {
		c := cors.New(cors.Options{})
		n.Use(c)
	}

	// add handler
	n.UseHandler(r)

	return n
}

func main() {
	// listen
	http.ListenAndServe(fmt.Sprintf(":%s", config.Port), App())
}
