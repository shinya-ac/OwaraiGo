FROM golang:1.19

WORKDIR /var/www

COPY ./src /var/www

CMD ["go","run","main.go"]