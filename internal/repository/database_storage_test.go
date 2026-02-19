package repository

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSuccessRepositoryPing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	rep := DataBase{database: db}
	mock.ExpectPing()
	errPing := rep.Ping()
	assert.NoError(t, errPing)
}

func TestNegativeRepositoryPing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)
	defer db.Close()

	rep := DataBase{database: db}
	mock.ExpectPing().WillReturnError(fmt.Errorf("ping error"))
	errPing := rep.Ping()
	if assert.NotNil(t, errPing) {
		assert.Contains(t, errPing.Error(), "ping error")
	}
}

func TestSuccessGetCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	name := uuid.New().String()
	cnt := rand.Int64()
	exptectedMetrics := models.Metrics{ID: name, MType: models.Counter, Delta: &cnt}
	q := regexp.QuoteMeta(`SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2`)

	mock.ExpectQuery(q).WithArgs(name, models.Counter).WillReturnRows(
		sqlmock.NewRows([]string{"name", "metric_type", "delta"}).AddRow(name, models.Counter, cnt))
	metric, err := rep.GetCounter(name)
	assert.NoError(t, err)
	assert.Equal(t, exptectedMetrics, *metric)
}

func TestNegativeGetCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	name := uuid.New().String()
	q := regexp.QuoteMeta(`SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2`)
	mock.ExpectQuery(q).WithArgs(name, models.Counter).WillReturnError(fmt.Errorf("sql error"))
	_, err = rep.GetCounter(name)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "sql error")
	}
}

func TestSuccessGetGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	name := uuid.New().String()
	cnt := rand.Float64()
	exptectedMetrics := models.Metrics{ID: name, MType: models.Counter, Value: &cnt}
	q := regexp.QuoteMeta(`SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2`)

	mock.ExpectQuery(q).WithArgs(name, models.Gauge).WillReturnRows(
		sqlmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Counter, cnt))
	metric, err := rep.GetGauge(name)
	assert.NoError(t, err)
	assert.Equal(t, exptectedMetrics, *metric)
}

func TestNegativeGetGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	name := uuid.New().String()
	q := regexp.QuoteMeta(`SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2`)

	mock.ExpectQuery(q).WithArgs(name, models.Gauge).WillReturnError(fmt.Errorf("sql error"))
	_, err = rep.GetGauge(name)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "sql error")
	}
}

func TestSuccessGetAllCounters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	q := regexp.QuoteMeta(`SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1`)

	name := uuid.New().String()
	cnt := rand.Int64()
	expectedMetrics := map[string]*models.Metrics{name: {ID: name, MType: models.Counter, Delta: &cnt}}
	mock.ExpectQuery(q).WithArgs(models.Counter).WillReturnRows(
		sqlmock.NewRows([]string{"name", "metric_type", "delta"}).AddRow(name, models.Counter, cnt))
	actualMetrics, errGetAll := rep.GetAllCounters()
	assert.NoError(t, errGetAll)
	assert.Equal(t, expectedMetrics, actualMetrics)
}

func TestErrorGetAllCounters(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()
	rep := DataBase{database: db}
	q := regexp.QuoteMeta(`SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1`)

	mock.ExpectQuery(q).WithArgs(models.Counter).WillReturnError(fmt.Errorf("sql error"))
	actualMetrics, errGetAll := rep.GetAllCounters()
	assert.Empty(t, actualMetrics)
	if assert.NotNil(t, errGetAll, "ожидалась ошибка") {
		assert.Contains(t, errGetAll.Error(), "sql error")
	}
}

func TestSuccessGetAllGauges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}
	q := regexp.QuoteMeta(`SELECT name, metric_type, value FROM metrics WHERE metric_type = $1`)

	name := uuid.New().String()
	val := rand.Float64()
	expectedMetrics := map[string]*models.Metrics{name: {ID: name, MType: models.Gauge, Value: &val}}
	mock.ExpectQuery(q).WithArgs(models.Gauge).WillReturnRows(
		sqlmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Gauge, val))
	actualMetrics, errGetAll := rep.GetAllGauges()
	assert.NoError(t, errGetAll)
	assert.Equal(t, expectedMetrics, actualMetrics)
}

func TestErrorGetAllGauges(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()
	rep := DataBase{database: db}
	q := regexp.QuoteMeta(`SELECT name, metric_type, value FROM metrics WHERE metric_type = $1`)

	mock.ExpectQuery(q).WithArgs(models.Gauge).WillReturnError(fmt.Errorf("sql error"))
	actualMetrics, errGetAll := rep.GetAllGauges()
	assert.Empty(t, actualMetrics)
	if assert.NotNil(t, errGetAll, "ожидалась ошибка") {
		assert.Contains(t, errGetAll.Error(), "sql error")
	}
}

func TestSuccessInsertUpdateGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()

	rep := DataBase{database: db}
	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	val := rand.Float64()
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Gauge, Value: &val}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, nil, *expectedMetric.Value).WillReturnResult(
		sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	errUpdate := rep.Update(slog.Default(), &expectedMetric)
	assert.NoError(t, errUpdate)
}

func TestNegativeInsertUpdateGauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()

	rep := DataBase{database: db}
	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	val := rand.Float64()
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Gauge, Value: &val}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, nil, *expectedMetric.Value).WillReturnError(fmt.Errorf("sql error"))
	mock.ExpectRollback()
	errUpdate := rep.Update(slog.Default(), &expectedMetric)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}

func TestSuccessInsertCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()

	rep := DataBase{database: db}
	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	cnt := rand.Int64()
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, *expectedMetric.Delta, nil).WillReturnResult(
		sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	errUpdate := rep.Update(slog.Default(), &expectedMetric)
	assert.NoError(t, errUpdate)
}

func TestNegativeInsertUpdateCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "ошибка создания мока")
	defer db.Close()

	rep := DataBase{database: db}
	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	cnt := rand.Int64()
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, *expectedMetric.Delta, nil).WillReturnError(fmt.Errorf("sql error"))
	mock.ExpectRollback()
	errUpdate := rep.Update(slog.Default(), &expectedMetric)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}

func TestSuccessAllUpdateType(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}

	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	cnt, val := rand.Int64(), rand.Float64()
	expectedMetrics := []*models.Metrics{
		{ID: uuid.New().String(), MType: models.Gauge, Delta: nil, Value: &val},
		{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt, Value: nil},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(q)
	mock.ExpectExec(q).WithArgs(expectedMetrics[0].ID, expectedMetrics[0].MType, expectedMetrics[0].Delta, *expectedMetrics[0].Value).WillReturnResult(
		sqlmock.NewResult(1, 1))
	mock.ExpectExec(q).WithArgs(expectedMetrics[1].ID, expectedMetrics[1].MType, *expectedMetrics[1].Delta, expectedMetrics[1].Value).WillReturnResult(
		sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	errUpdate := rep.Updates(slog.Default(), expectedMetrics)
	assert.NoError(t, errUpdate)
}

func TestNegativeAllUpdateType(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	rep := DataBase{database: db}

	q := regexp.QuoteMeta(
		`INSERT INTO metrics (name, metric_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name, metric_type)
		DO UPDATE SET
	  	delta = CASE
			WHEN EXCLUDED.delta IS NOT NULL
		  	THEN COALESCE(metrics.delta, 0) + EXCLUDED.delta
			ELSE metrics.delta
	  	END,
	  	value = CASE
			WHEN EXCLUDED.value IS NOT NULL
		  	THEN EXCLUDED.value
			ELSE metrics.value
	  	END;`)

	cnt, val := rand.Int64(), rand.Float64()
	expectedMetrics := []*models.Metrics{
		{ID: uuid.New().String(), MType: models.Gauge, Delta: nil, Value: &val},
		{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt, Value: nil},
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(q)
	mock.ExpectExec(q).WithArgs(expectedMetrics[0].ID, expectedMetrics[0].MType, expectedMetrics[0].Delta, *expectedMetrics[0].Value).WillReturnResult(
		sqlmock.NewResult(1, 1))
	mock.ExpectExec(q).WithArgs(expectedMetrics[1].ID, expectedMetrics[1].MType, *expectedMetrics[1].Delta, expectedMetrics[1].Value).WillReturnError(
		fmt.Errorf("sql error"))
	mock.ExpectRollback()

	errUpdate := rep.Updates(slog.Default(), expectedMetrics)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}
