# go-metrics-collector

## Getting started

### env

For start docker you should create `.env` file (look into `.env.template` for more info).

For server you should create `.env.server` and for agent you should create `.env.agent` files.

### Running apps

About agent and server read more information in **cmd** folder.

### Docker

```bash
$ docker-compose up -d
```

## Changelog

You can find changelog in `/internal/README.md` file.

## Template update

Add template in remote:

```bash
$ git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

For autotests code update:

```bash 
$ git fetch template 
$ git checkout template/main .github
```

[README autotests](https://github.com/Yandex-Practicum/go-autotests).
