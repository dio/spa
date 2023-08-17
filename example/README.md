Do the following to run the example:

```console
$ cd client
$ yarn && PUBLIC_URL=%DEPLOYMENT_PATH% yarn build
$ cd ..
$ go run main.go # then go to the browser: http://localhost:3000/ok. Note that refreshing the browser on http://localhost:3000/ok/next works.
```
