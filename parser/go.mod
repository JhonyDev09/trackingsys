module github.com/JhonyDev09/trackingsys/parser

go 1.26.4

require (
	github.com/JhonyDev09/trackingsys/shared v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.7.2
	github.com/rabbitmq/amqp091-go v1.12.0
)

replace github.com/JhonyDev09/trackingsys/shared => ../shared

replace golang.org/x/crypto => github.com/golang/crypto v0.31.0

replace golang.org/x/text => github.com/golang/text v0.21.0

replace golang.org/x/sync => github.com/golang/sync v0.10.0
