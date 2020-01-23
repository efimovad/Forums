package forum_rep

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/models"
	"github.com/lib/pq"

	//"github.com/go-openapi/strfmt"
	"strconv"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewForumRepository(db *sql.DB) forum.Repository {
	return &Repository{ db}
}

func (r *Repository) CreateForum(forum *models.Forum) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	err = tx.QueryRow(
		"INSERT INTO forums (slug, title, \"user\") VALUES ($1, $2, $3) RETURNING id",
		forum.Slug,
		forum.Title,
		forum.User,
	).Scan(&forum.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) CreateThread(thread *models.Thread) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if thread.Created.IsZero() {
		thread.Created = time.Now()
	}

	err = tx.QueryRow("INSERT INTO threads (forum, author, created, message, title, slug, votes) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		thread.Forum,
		thread.Author,
		thread.Created,
		thread.Message,
		thread.Title,
		thread.Slug,
		thread.Votes,
	).Scan(&thread.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindBySlug(slug string) (*models.Forum, error) {
	f := new(models.Forum)

	if err := r.db.QueryRow(
		"SELECT id, slug, title, \"user\", " +
			"(SELECT COUNT(*) FROM threads WHERE LOWER(forum) = LOWER($1)), " +
			"posts " +
			"FROM forums WHERE LOWER(slug) = LOWER($1)",
		slug,
	).Scan(
		&f.ID,
		&f.Slug,
		&f.Title,
		&f.User,
		&f.Threads,
		&f.Posts,
	); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *Repository) GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error){
	var err error
	var rows *sql.Rows
	var threads []*models.Thread
	var t time.Time
	var sinceSet bool

	if params.Since != "" {
		layout := "2006-01-02T15:04:05Z07:00"
		t, err = time.Parse(layout, params.Since)
		if err != nil {
			return nil, err
		}
		sinceSet = true
	}

	rows, err = r.db.Query(
		`SELECT id, forum, author, created, message, title, slug, votes 
						FROM threads
						WHERE LOWER(forum) = LOWER($1) AND 
						      (NOT $5 OR (NOT $3 AND created >= $2) OR ($3 AND created <= $2))
						ORDER BY
							CASE WHEN $3 THEN created END DESC,
							CASE WHEN NOT $3 THEN created END ASC
						LIMIT CASE WHEN $4 > 0 THEN $4 END;`,
		slug, t, params.Desc, params.Limit, sinceSet)


	if err != nil {
		return nil, err
	}

	for rows.Next() {
		t := new(models.Thread)
		err := rows.Scan(&t.ID, &t.Forum, &t.Author, &t.Created, &t.Message, &t.Title, &t.Slug, &t.Votes)
		if err != nil {
			return nil, err
		}
		threads = append(threads, t)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return threads, nil
}

func (r *Repository) FindThread(id int64) (*models.Thread, error) {
	t := new(models.Thread)
	if err := r.db.QueryRow(
		"SELECT id, forum, author, created, message, title, slug, votes FROM threads WHERE id = $1",
		id,
	).Scan(
		&t.ID,
		&t.Forum,
		&t.Author,
		&t.Created,
		&t.Message,
		&t.Title,
		&t.Slug,
		&t.Votes,
	); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) FindThreadBySlug(slug string) (*models.Thread, error) {
	t := new(models.Thread)
	if err := r.db.QueryRow(
		"SELECT id, forum, author, created, message, title, slug, votes FROM threads " +
			"WHERE LOWER(slug) = LOWER($1)",
		slug,
	).Scan(
		&t.ID,
		&t.Forum,
		&t.Author,
		&t.Created,
		&t.Message,
		&t.Title,
		&t.Slug,
		&t.Votes,
	); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) UpdateThread(thread *models.Thread) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE threads SET votes = $1, title = $2, message = $3 WHERE id = $4")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(thread.Votes, thread.Title, thread.Message, thread.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r * Repository) CreatePosts(posts []*models.Post, thread *models.Thread) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	now := time.Now()

	sqlStr := `
		INSERT INTO posts(id, parent, thread, forum, author, created, message, path)
			VALUES
	`

	vals := []interface{}{}
	for _, post := range posts {
		var author string
		err = r.db.QueryRow(`
			SELECT nickname
				FROM users
				WHERE LOWER(nickname) = LOWER($1)
			`,
			post.Author,
		).Scan(&author)
		// Если хотя бы одного юзера не существует - откатываемся
		if err != nil || author == "" {
			_ = tx.Rollback()
			return errors.New(forum.NOT_FOUND_ERR)
		}

		if post.Parent == 0 {
			// Создание массива пути с единственным значением -
			// id создаваемого сообщения
			sqlStr += "(nextval('posts_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"ARRAY[currval(pg_get_serial_sequence('posts', 'id'))::bigint]),"


			vals = append(vals, post.Parent, thread.ID, thread.Forum, post.Author, now, post.Message)
		} else {
			var parentThreadId int64
			err = r.db.QueryRow(`
				SELECT thread
					FROM posts
					WHERE id = $1
				`,
				post.Parent,
			).Scan(
				&parentThreadId,
			)
			if err != nil {
				_ = tx.Rollback()
				return errors.New("Parent post was created in another thread")
			}

			if parentThreadId != thread.ID {
				_ = tx.Rollback()
				return errors.New("Parent post was created in another thread")
			}

			// Конкатенация 2-х массивов
			sqlStr += " (nextval('posts_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"(SELECT path FROM posts WHERE id = ? AND thread = ?) || " +
				"currval(pg_get_serial_sequence('posts', 'id'))::bigint),"

			vals = append(vals, post.Parent, thread.ID, thread.Forum, post.Author, now, post.Message, post.Parent, thread.ID)
		}

	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	sqlStr += `
		RETURNING id, parent, thread, forum, author, created, message, isEdited 
	`

	sqlStr = ReplaceSQL(sqlStr, "?")
	if len(posts) > 0 {
		rows, err := tx.Query(sqlStr, vals...)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		i := 0
		for rows.Next() {
			err := rows.Scan(
				&(posts[i]).ID,
				&(posts)[i].Parent,
				&(posts)[i].Thread,
				&(posts)[i].Forum,
				&(posts)[i].Author,
				&(posts)[i].Created,
				&(posts)[i].Message,
				&(posts)[i].IsEdited,
			)
			i += 1


			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	_, err = tx.Exec(`
		UPDATE forums
			SET posts = posts + $1
			WHERE lower(slug) = lower($2)
		`,
		len(posts),
		thread.Forum,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) FindPost(id int64) (*models.Post, error) {
	p := new(models.Post)
	if err := r.db.QueryRow(
		"SELECT id, author, created, forum, isEdited, message, parent, thread FROM posts WHERE id = $1",
		id,
	).Scan(
		&p.ID,
		&p.Author,
		&p.Created,
		&p.Forum,
		&p.IsEdited,
		&p.Message,
		&p.Parent,
		&p.Thread,
	); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) UpdatePost(post *models.Post) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE posts SET message = $1, isEdited = $2 WHERE id = $3 RETURNING id")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(post.Message, post.IsEdited, post.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CreateVote(vote *models.Vote, thread *models.Thread) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare("INSERT INTO votes(nickname, vote, thread) VALUES ($1, $2, $3) ON CONFLICT (LOWER(nickname), thread) DO UPDATE SET vote = $2")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(vote.Nickname, vote.Voice, thread.ID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	var numVotes int64
	rowT := tx.QueryRow("SELECT votes FROM threads WHERE id = $1", thread.ID)
	err = rowT.Scan(
		&numVotes,
	)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return numVotes, nil
}

func (r *Repository) GetPosts(thread *models.Thread, params *models.ListParameters) ([]*models.Post, error) {
	posts := make([]*models.Post, 0)

	var query string

	conditionSign := ">"
	if params.Desc == true {
		conditionSign = "<"
	}

	order := "asc"
	if params.Desc {
		order = "desc"
	}

	if params.Sort == "flat" {
		query = "SELECT id, parent, thread, forum, author, created, message, isedited, path FROM posts WHERE thread = $1 "
		if params.Since != "" {
			query += fmt.Sprintf(" AND id %s %s ", conditionSign, params.Since)
		}
		query += fmt.Sprintf(" ORDER BY created %s, id %s LIMIT %d", order, order, params.Limit)
	} else if params.Sort == "tree" {
		orderString := fmt.Sprintf(" ORDER BY path[1] %s, path %s ", order, order)
		query = "SELECT id, parent, thread, forum, author, created, message, isedited, path " +
			"FROM posts " +
			"WHERE thread = $1 "
		if params.Since != "" {
			query += fmt.Sprintf(" AND path %s (SELECT path FROM posts WHERE id = %s) ", conditionSign, params.Since)
		}
		query += orderString
		query += fmt.Sprintf("LIMIT %d", params.Limit)
	}else if params.Sort == "parent_tree" {
		query = "SELECT id, parent, thread, forum, author, created, message, isedited, path " +
			"FROM posts " +
			"WHERE thread = $1 AND path && (SELECT ARRAY (select id from posts WHERE thread = $1 AND parent = 0 "
		if params.Since != "" {
			query += fmt.Sprintf(" AND path %s (SELECT path[1:1] FROM posts WHERE id = %s) ", conditionSign, params.Since)
		}
		query += fmt.Sprintf("ORDER BY path[1] %s, path LIMIT %d)) ", order, params.Limit)
		query += fmt.Sprintf("ORDER BY path[1] %s, path ", order)
	}

	rows, err := r.db.Query(query, thread.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		p := models.Post{}
		err := rows.Scan(&p.ID, &p.Parent, &p.Thread, &p.Forum, &p.Author, &p.Created, &p.Message, &p.IsEdited, pq.Array(&p.Path))
		if err != nil {
			return nil, err
		}

		posts = append(posts, &p)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *Repository) GetUsers(id int64, params models.ListParameters) ([]*models.User, error) {
	var err error
	var rows *sql.Rows
	var users []*models.User

	rows, err = r.db.Query(
		`SELECT nickname, fullname, about, email FROM users
    			WHERE id IN (SELECT user_id FROM forum_users WHERE forum_id = $1) AND 
    			      ($2 = '' OR (NOT $3 AND LOWER(nickname) > $2) OR ($3 AND LOWER(nickname) < $2))
				ORDER BY 
				         CASE WHEN NOT $3 THEN LOWER(nickname) END,
				         CASE WHEN $3 THEN LOWER(nickname) END DESC
				LIMIT CASE WHEN $4 > 0 THEN $4 END;`,
		id, strings.ToLower(params.Since), params.Desc, params.Limit)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		item := new(models.User)

		err := rows.Scan(&item.Nickname, &item.FullName, &item.About, &item.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, item)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) FindUser(nickname string) (*models.User, error) {
	u := new(models.User)
	if err := r.db.QueryRow(
		"SELECT id, email, about, fullname, nickname FROM users WHERE LOWER(nickname) = LOWER($1)",
		nickname,
	).Scan(
		&u.ID,
		&u.Email,
		&u.About,
		&u.FullName,
		&u.Nickname,
	); err != nil {
		return nil, err
	}
	return u, nil
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}