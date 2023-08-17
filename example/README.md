Do the following to run the example:

```console
$ cd client
$ yarn && PUBLIC_URL=%DEPLOYMENT_PATH% yarn build
$ cd ..
$ # with prefix "ok"
$ go run main.go -prefix ok # then go to the browser: http://localhost:3000/ok. Note that refreshing the browser on http://localhost:3000/ok/next works.
$ # without prefix
$ go run main.go # then go to the browser: http://localhost:3000. Note that refreshing the browser on http://localhost:3000/ works.
```
