FROM golang:1.19

RUN mkdir /var/www
COPY main.go /var/www

RUN echo "hello wrold" >> /var/www/sample.txt

CMD ["go", "run", "/var/www/main.go"]