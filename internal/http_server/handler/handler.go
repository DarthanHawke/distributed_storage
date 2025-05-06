package handler

import (
	"distributed_storage/internal/logger/slogerror"
	trlog "distributed_storage/internal/logger/transaction_logger"
	"distributed_storage/internal/storage"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Request struct {
	DataValue string `json:"value"`
}

// keyValuePutHandler ожидает получить PUT-запрос с ресурсом "/v1/key/{key}".
func NewPutHandler(transactLog *trlog.FileTransactionLogger, log *slog.Logger, dataPutting storage.StoreUsage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.NewPutHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		key := chi.URLParam(r, "key")

		value, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read request body", slogerror.Error(err))
			render.JSON(w, r, "faild to read request")
			return
		}
		defer r.Body.Close()

		log.Info("request body readed", slog.Any("request", value))

		err = dataPutting.PutData(key, string(value)) // Сохранить значение как строку
		if err != nil {
			log.Error("faild to add url", slogerror.Error(err))
			render.JSON(w, r, "faild to add value")
			return
		}
		transactLog.WritePut(key, string(value)) // Сохранить значение как строку

		w.WriteHeader(http.StatusCreated) // Все хорошо! Вернуть StatusCreated

		log.Info("PUT key=%s value=%s\n", key, string(value))
	}
}

func NewGetHandler(log *slog.Logger, dataGetting storage.StoreUsage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.NewGetHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		key := chi.URLParam(r, "key")

		value, err := dataGetting.GetData(key) // Получить значение для данного ключа
		if errors.Is(err, storage.ErrorNoSuchKey) {
			log.Error("key not exist", slogerror.Error(err))
			render.JSON(w, r, "key not exist")
			return
		}
		if err != nil {
			log.Error("error of getting data", slogerror.Error(err))
			render.JSON(w, r, "error of getting data")
			return
		}
		w.Write([]byte(value)) // Записать значение в ответ

		slog.String("GET key=%s\n", key)

	}
}

func NewDeleteHandler(transactLog *trlog.FileTransactionLogger, log *slog.Logger, DataDeleting storage.StoreUsage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.NewDeleteHandler"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		key := chi.URLParam(r, "key")

		err := DataDeleting.DeleteData(key) // Получить значение для данного ключа
		if errors.Is(err, storage.ErrorNoSuchKey) {
			log.Error("key not exist", slogerror.Error(err))
			render.JSON(w, r, "key not exist")
			return
		}
		if err != nil {
			log.Error("error of deleting data", slogerror.Error(err))
			render.JSON(w, r, "error of deleting data")
			return
		}

		transactLog.WriteDelete(key) // Сохранить значение как строку

		slog.String("DELETE key=%s\n", key)
	}
}
