FROM golang:1.13.4-stretch AS build

WORKDIR /usr/src/tech-db

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM ubuntu:18.04 AS release


ENV PGVER 10
RUN apt -y update && apt install -y postgresql-$PGVER

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "include_dir='conf.d'" >> /etc/postgresql/$PGVER/main/postgresql.conf
ADD ./postgresql.conf /etc/postgresql/$PGVER/main/conf.d/basic.conf

EXPOSE 5432

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

EXPOSE 5000

COPY internal/store/functions.sql .
COPY --from=build /usr/src/tech-db/forum .

CMD service postgresql start && ./forum --scheme=http --port=5000 --host=0.0.0.0 --database=postgres://docker:docker@localhost/docker