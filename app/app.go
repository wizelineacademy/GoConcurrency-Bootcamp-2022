package app

import "GoConcurrency-Bootcamp-2022/router"

func Start() {
	r := router.Init()
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
