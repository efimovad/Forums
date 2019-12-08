FROM golang

WORKDIR /Documents/technopark/tech_sem2/databases/workspace

COPY go.mod .
#COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM ubuntu:19.04
ENV DEBIAN_FRONTEND=noninteractive
ENV PGVER 11
ENV PORT 5000
ENV POSTGRES_HOST localhost
ENV POSTGRES_PORT 5432
ENV POSTGRES_DB docker
ENV POSTGRES_USER docker
ENV POSTGRES_PASSWORD docker
EXPOSE $PORT

RUN apt-get update && apt-get install -y postgresql-$PGVER

USER postgres

RUN service postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    psql --command "CREATE DATABASE docker OWNER docker;" &&\
    #createdb -O docker docker &&\
    service postgresql stop

#COPY config/pg_hba.conf /etc/postgresql/$PGVER/main/pg_hba.conf
#COPY config/postgresql.conf /etc/postgresql/$PGVER/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

#COPY sunrise_db.sql .
#COPY --from=builder /usr/src/app/DB_TP .
COPY ./forum .

CMD service postgresql start && ./forum