# Steps
## Postgres Setup
Create postgres container using docker:
docker run --name postgres -e POSTGRES_PASSWORD=postgres -d postgres

Open postgres shell in container and create Table
```
docker exec -it postgres psql -U postgres
CREATE TABLE products(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
);
```

## Setup Github Repo
Create repo in github:
<img width="500" alt="Bildschirm­foto 2023-04-06 um 11 25 53" src="https://user-images.githubusercontent.com/30144387/230335436-25e7583d-999e-4d57-93f0-71cb0f1c3526.png">

Clone repo:


\end{document}
