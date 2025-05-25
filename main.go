package main

import (
	"net/http"

	"github.com/angelbarreiros/Penguin/router"
	"github.com/angelbarreiros/Penguin/router/handlers"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	var i18n = handlers.GetI18nInstance(handlers.SetDefaultLocale("en"), handlers.SetDirectory("i18n"))
	var translated = i18n.Translate("INTERNAL_ERROR", r)

	w.Write([]byte(translated))

}

func main() {
	r := router.Router()
	r.NewRoute(router.Route{
		Path:             "/",
		Method:           "GET",
		Handler:          handlerUser,
		AditionalMethods: []router.HTTPMethod{router.HEAD, router.OPTIONS},
	})

	r.StartServer(":8080")

}
