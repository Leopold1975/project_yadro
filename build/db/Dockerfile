FROM docker.io/library/postgres:15

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=qwerty

COPY ./.init/init.sql /docker-entrypoint-initdb.d/init.sql

EXPOSE 5432