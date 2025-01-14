package mysql

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/usememos/memos/store"
)

func (d *DB) CreateUser(ctx context.Context, create *store.User) (*store.User, error) {
	fields := []string{"`username`", "`role`", "`email`", "`nickname`", "`password_hash`", "`avatar_url`"}
	placeholder := []string{"?", "?", "?", "?", "?", "?"}
	args := []any{create.Username, create.Role, create.Email, create.Nickname, create.PasswordHash, create.AvatarURL}

	if create.RowStatus != "" {
		fields = append(fields, "`row_status`")
		placeholder = append(placeholder, "?")
		args = append(args, create.RowStatus)
	}

	if create.CreatedTs != 0 {
		fields = append(fields, "`created_ts`")
		placeholder = append(placeholder, "FROM_UNIXTIME(?)")
		args = append(args, create.CreatedTs)
	}

	if create.UpdatedTs != 0 {
		fields = append(fields, "`updated_ts`")
		placeholder = append(placeholder, "FROM_UNIXTIME(?)")
		args = append(args, create.UpdatedTs)
	}

	if create.ID != 0 {
		fields = append(fields, "`id`")
		placeholder = append(placeholder, "?")
		args = append(args, create.ID)
	}

	stmt := "INSERT INTO user (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ")"
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	id64 := int32(id)
	list, err := d.ListUsers(ctx, &store.FindUser{ID: &id64})
	if err != nil {
		return nil, err
	}
	if len(list) != 1 {
		return nil, errors.Wrapf(nil, "unexpected user count: %d", len(list))
	}

	return list[0], nil
}

func (d *DB) UpdateUser(ctx context.Context, update *store.UpdateUser) (*store.User, error) {
	set, args := []string{}, []any{}
	if v := update.UpdatedTs; v != nil {
		set, args = append(set, "`updated_ts` = ?"), append(args, *v)
	}
	if v := update.RowStatus; v != nil {
		set, args = append(set, "`row_status` = ?"), append(args, *v)
	}
	if v := update.Username; v != nil {
		set, args = append(set, "`username` = ?"), append(args, *v)
	}
	if v := update.Email; v != nil {
		set, args = append(set, "`email` = ?"), append(args, *v)
	}
	if v := update.Nickname; v != nil {
		set, args = append(set, "`nickname` = ?"), append(args, *v)
	}
	if v := update.AvatarURL; v != nil {
		set, args = append(set, "`avatar_url` = ?"), append(args, *v)
	}
	if v := update.PasswordHash; v != nil {
		set, args = append(set, "`password_hash` = ?"), append(args, *v)
	}
	args = append(args, update.ID)

	query := "UPDATE `user` SET " + strings.Join(set, ", ") + " WHERE `id` = ?"
	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}

	user := &store.User{}
	query = "SELECT `id`, `username`, `role`, `email`, `nickname`, `password_hash`, `avatar_url`, UNIX_TIMESTAMP(`created_ts`), UNIX_TIMESTAMP(`updated_ts`), `row_status` FROM `user` WHERE `id` = ?"
	if err := d.db.QueryRowContext(ctx, query, update.ID).Scan(
		&user.ID,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.Nickname,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.CreatedTs,
		&user.UpdatedTs,
		&user.RowStatus,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (d *DB) ListUsers(ctx context.Context, find *store.FindUser) ([]*store.User, error) {
	where, args := []string{"1 = 1"}, []any{}

	if v := find.ID; v != nil {
		where, args = append(where, "`id` = ?"), append(args, *v)
	}
	if v := find.Username; v != nil {
		where, args = append(where, "`username` = ?"), append(args, *v)
	}
	if v := find.Role; v != nil {
		where, args = append(where, "`role` = ?"), append(args, *v)
	}
	if v := find.Email; v != nil {
		where, args = append(where, "`email` = ?"), append(args, *v)
	}
	if v := find.Nickname; v != nil {
		where, args = append(where, "`nickname` = ?"), append(args, *v)
	}

	query := "SELECT `id`, `username`, `role`, `email`, `nickname`, `password_hash`, `avatar_url`, UNIX_TIMESTAMP(`created_ts`), UNIX_TIMESTAMP(`updated_ts`), `row_status` FROM `user` WHERE " + strings.Join(where, " AND ") + " ORDER BY `created_ts` DESC, `row_status` DESC"
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*store.User, 0)
	for rows.Next() {
		var user store.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&user.Email,
			&user.Nickname,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.CreatedTs,
			&user.UpdatedTs,
			&user.RowStatus,
		); err != nil {
			return nil, err
		}
		list = append(list, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (d *DB) DeleteUser(ctx context.Context, delete *store.DeleteUser) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM `user` WHERE `id` = ?", delete.ID)
	if err != nil {
		return err
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}

	if err := d.Vacuum(ctx); err != nil {
		// Prevent linter warning.
		return err
	}

	return nil
}
