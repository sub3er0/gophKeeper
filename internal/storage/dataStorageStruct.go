package storage

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// DBConnectionInterface определяет методы для взаимодействия с базой данных.
// Этот интерфейс позволяет выполнять запросы, отправлять команды,
// проверять соединение и закрывать соединения.
type DBConnectionInterface interface {
	// Query выполняет SQL-запрос и возвращает строки результата.
	// Принятый контекст позволяет отменять запросы.
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// Exec выполняет SQL-команду и возвращает тег команды,
	// указывающий на количество затронутых строк.
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)

	// Ping проверяет состояние подключения к базе данных.
	// Возвращает ошибку, если соединение недоступно.
	Ping(ctx context.Context) error

	// Close закрывает соединение с базой данных.
	Close()

	// SendBatch отправляет пакет запросов в базу данных.
	// Возвращает результаты отправленных батчей.
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	// QueryRow возвращает одну строку результата
	QueryRow(context.Context, string, ...interface{}) pgx.Row

	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}
