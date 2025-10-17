# Docker Compose Example

Simple Docker Compose setup for running Dynapins Server.

## Quick Start

1. **Generate a private key:**

```bash
openssl ecparam -genkey -name prime256v1 -noout -out private_key.pem
```

2. **Set environment variable:**

```bash
export PRIVATE_KEY_PEM="$(cat private_key.pem)"
```

3. **Start the server:**

```bash
docker-compose up -d
```

4. **Test it:**

```bash
curl "http://localhost:8080/v1/pins?domain=example.com"
```

## Using .env File

1. Copy the template:

```bash
cp env-template.txt .env
```

2. Edit `.env` and add your private key

3. Uncomment `env_file` in `docker-compose.yml`

4. Start:

```bash
docker-compose up -d
```

## Logs

View logs:

```bash
docker-compose logs -f
```

## Stop

```bash
docker-compose down
```


