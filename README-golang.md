# Embed Next.Js front-end application into Golang binary

## Add `export` script

Add the following into `package.json` 

```json
"scripts": {
    "export": "next export",
  }
```

## Build and Export Next.js front-end application

```shell
npm run build
npm run export
```

## Build Golang application

```shell
go build main.go
```

## Run 

```shell
./main
```

open `http://localhost:8080` in browser

## Export static files

```shell
./main --export `pwd`/t
```

## Setup overlay folder

```shell
./main --physical-root `pwd`/t
```

