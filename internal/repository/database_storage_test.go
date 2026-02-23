package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	models "github.com/serg1732/practicum-yandex-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSuccessRepositoryPing(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	mock.ExpectPing()
	errPing := rep.Ping(t.Context())
	assert.NoError(t, errPing)
}

func TestNegativeRepositoryPing(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	mock.ExpectPing().WillReturnError(fmt.Errorf("ping error"))
	errPing := rep.Ping(t.Context())
	if assert.NotNil(t, errPing) {
		assert.Contains(t, errPing.Error(), "ping error")
	}
}

func TestSuccessGetCounter(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	name := uuid.New().String()
	cnt := rand.Int64()
	exptectedMetrics := models.Metrics{ID: name, MType: models.Counter, Delta: &cnt}
	q := `SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2`

	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(name, models.Counter).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "delta"}).AddRow(name, models.Counter, &cnt))
	mock.ExpectCommit()
	metric, err := rep.GetCounter(t.Context(), name)
	assert.NoError(t, err)
	assert.Equal(t, exptectedMetrics, *metric)
}

func TestNegativeGetCounter(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	name := uuid.New().String()
	cnt := rand.Int64()
	q := `SELECT name, metric_type, delta FROM metrics WHERE name = $1 AND metric_type = $2`
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(name, models.Counter).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "delta"}).AddRow(name, models.Counter, &cnt).RowError(
			0, errors.New("sql error")))
	mock.ExpectRollback()
	_, err = rep.GetCounter(t.Context(), name)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "sql error")
	}
}

func TestSuccessGetGauge(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	name := uuid.New().String()
	gauge := rand.Float64()
	exptectedMetrics := models.Metrics{ID: name, MType: models.Counter, Value: &gauge}
	q := `SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2`

	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(name, models.Gauge).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Counter, &gauge))
	mock.ExpectCommit()
	metric, err := rep.GetGauge(t.Context(), name)
	assert.NoError(t, err)
	assert.Equal(t, exptectedMetrics, *metric)
}

func TestNegativeGetGauge(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	name := uuid.New().String()
	gauge := rand.Float64()
	q := `SELECT name, metric_type, value FROM metrics WHERE name = $1 AND metric_type = $2`

	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(name, models.Gauge).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Counter, &gauge).
			RowError(0, errors.New("sql error")))
	mock.ExpectRollback()
	_, err = rep.GetGauge(t.Context(), name)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "sql error")
	}
}

func TestSuccessGetAllCounters(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1`

	name := uuid.New().String()
	cnt := rand.Int64()
	expectedMetrics := map[string]*models.Metrics{name: {ID: name, MType: models.Counter, Delta: &cnt}}
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(models.Counter).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "delta"}).AddRow(name, models.Counter, &cnt))
	mock.ExpectCommit()
	actualMetrics, errGetAll := rep.GetAllCounters(t.Context())
	assert.NoError(t, errGetAll)
	assert.Equal(t, expectedMetrics, actualMetrics)
}

func TestErrorGetAllCounters(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `SELECT name, metric_type, delta FROM metrics WHERE metric_type = $1`

	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(models.Counter).WillReturnError(fmt.Errorf("sql error"))
	actualMetrics, errGetAll := rep.GetAllCounters(t.Context())
	mock.ExpectRollback()
	assert.Empty(t, actualMetrics)
	if assert.NotNil(t, errGetAll, "ожидалась ошибка") {
		assert.Contains(t, errGetAll.Error(), "sql error")
	}
}

func TestSuccessGetAllGauges(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `SELECT name, metric_type, value FROM metrics WHERE metric_type = $1`

	name := uuid.New().String()
	val := rand.Float64()
	expectedMetrics := map[string]*models.Metrics{name: {ID: name, MType: models.Gauge, Value: &val}}
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(models.Gauge).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Gauge, &val))
	mock.ExpectCommit()
	actualMetrics, errGetAll := rep.GetAllGauges(t.Context())
	assert.NoError(t, errGetAll)
	assert.Equal(t, expectedMetrics, actualMetrics)
}

func TestErrorGetAllGauges(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `SELECT name, metric_type, value FROM metrics WHERE metric_type = $1`
	name := uuid.New().String()
	val := rand.Float64()
	mock.ExpectBegin()
	mock.ExpectQuery(q).WithArgs(models.Gauge).WillReturnRows(
		pgxmock.NewRows([]string{"name", "metric_type", "value"}).AddRow(name, models.Gauge, &val).
			RowError(0, errors.New("sql error")))
	mock.ExpectRollback()
	actualMetrics, errGetAll := rep.GetAllGauges(t.Context())
	assert.Empty(t, actualMetrics)
	if assert.NotNil(t, errGetAll, "ожидалась ошибка") {
		assert.Contains(t, errGetAll.Error(), "sql error")
	}
}

func TestSuccessInsertUpdateGauge(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `INSERT INTO metrics (name, metric_type, delta, value)
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
	  	END;`

	val := rand.Float64()
	var delta *int64
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Gauge, Value: &val}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, delta, expectedMetric.Value).WillReturnResult(
		pgxmock.NewResult("1", 1))
	mock.ExpectCommit()
	errUpdate := rep.Update(t.Context(), slog.Default(), &expectedMetric)
	assert.NoError(t, errUpdate)
}

func TestNegativeInsertUpdateGauge(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
	q := `INSERT INTO metrics (name, metric_type, delta, value)
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
	  	END;`

	val := rand.Float64()
	var delta *int64
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Gauge, Value: &val}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, delta, expectedMetric.Value).WillReturnError(fmt.Errorf("sql error"))
	mock.ExpectRollback()
	errUpdate := rep.Update(t.Context(), slog.Default(), &expectedMetric)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}

func TestSuccessInsertCounter(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
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
	var val *float64
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, expectedMetric.Delta, val).WillReturnResult(
		pgxmock.NewResult("1", 1))
	mock.ExpectCommit()
	errUpdate := rep.Update(t.Context(), slog.Default(), &expectedMetric)
	assert.NoError(t, errUpdate)
}

func TestNegativeInsertUpdateCounter(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}
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
	var val *float64
	expectedMetric := models.Metrics{ID: uuid.New().String(), MType: models.Counter, Delta: &cnt}
	mock.ExpectBegin()
	mock.ExpectExec(q).WithArgs(expectedMetric.ID, expectedMetric.MType, expectedMetric.Delta, val).WillReturnError(fmt.Errorf("sql error"))
	mock.ExpectRollback()
	errUpdate := rep.Update(t.Context(), slog.Default(), &expectedMetric)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}

func TestSuccessAllUpdateType(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}

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

	const prepareName string = "insert_metrics"
	mock.ExpectBegin()
	mock.ExpectPrepare(prepareName, q)
	mock.ExpectExec(prepareName).WithArgs(expectedMetrics[0].ID, expectedMetrics[0].MType, expectedMetrics[0].Delta, expectedMetrics[0].Value).WillReturnResult(
		pgxmock.NewResult("1", 1))
	mock.ExpectExec(prepareName).WithArgs(expectedMetrics[1].ID, expectedMetrics[1].MType, expectedMetrics[1].Delta, expectedMetrics[1].Value).WillReturnResult(
		pgxmock.NewResult("2", 1))
	mock.ExpectCommit()

	errUpdate := rep.Updates(t.Context(), slog.Default(), expectedMetrics)
	assert.NoError(t, errUpdate)
}

func TestNegativeAllUpdateType(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	rep := DataBase{database: mock}

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
	const prepareName string = "insert_metrics"
	mock.ExpectPrepare(prepareName, q)
	mock.ExpectExec(prepareName).WithArgs(expectedMetrics[0].ID, expectedMetrics[0].MType, expectedMetrics[0].Delta, expectedMetrics[0].Value).WillReturnResult(
		pgxmock.NewResult("1", 1))
	mock.ExpectExec(prepareName).WithArgs(expectedMetrics[1].ID, expectedMetrics[1].MType, expectedMetrics[1].Delta, expectedMetrics[1].Value).WillReturnError(
		fmt.Errorf("sql error"))
	mock.ExpectRollback()

	errUpdate := rep.Updates(t.Context(), slog.Default(), expectedMetrics)
	if assert.NotNil(t, errUpdate) {
		assert.Contains(t, errUpdate.Error(), "sql error")
	}
}
