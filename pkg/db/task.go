package db

import (
	"database/sql"
	"fmt"
)

// Tasks возвращает список задач, отсортированных по дате. limit - максимальное количество возвращаемых задач.
// search - опциональная строка поиска.
func Tasks(limit int, search string) ([]*Task, error) {
	var query string
	var args []interface{}

	// Базовый запрос
	baseQuery := "SELECT id, date, title, comment, repeat FROM scheduler"

	if search != "" {
		if len(search) == 10 && search[2] == '.' && search[5] == '.' {
			day := search[0:2]
			month := search[3:5]
			year := search[6:10]
			dbDate := year + month + day

			query = baseQuery + " WHERE date = ? ORDER BY date ASC LIMIT ?"
			args = append(args, dbDate, limit)
		} else {
			searchPattern := "%" + search + "%"
			query = baseQuery + " WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?"
			args = append(args, searchPattern, searchPattern, limit)
		}
	} else {
		query = baseQuery + " ORDER BY date ASC LIMIT ?"
		args = append(args, limit)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		var idInt int64
		err := rows.Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		t.ID = fmt.Sprintf("%d", idInt)
		tasks = append(tasks, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if tasks == nil {
		tasks = make([]*Task, 0)
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	var t Task
	var idInt int64

	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	t.ID = fmt.Sprintf("%d", idInt)
	return &t, nil
}

// UpdateTask обновляет существующую задачу
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found or no changes made")
	}

	return nil
}

func DeleteTask(id string) error {
	query := "DELETE FROM scheduler WHERE id = ?"
	res, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// UpdateDate обновляет только дату задачи
func UpdateDate(id string, newDate string) error {
	query := "UPDATE scheduler SET date = ? WHERE id = ?"
	res, err := DB.Exec(query, newDate, id)
	if err != nil {
		return fmt.Errorf("failed to update date: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
