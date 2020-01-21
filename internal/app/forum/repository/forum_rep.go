package forum_rep

import (
	"database/sql"
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/models"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	MaxPostNum = 6
)

type Repository struct {
	db *sql.DB
}

func NewForumRepository(db *sql.DB) forum.Repository {
	return &Repository{ db}
}

func (r *Repository) CreateForum(forum *models.Forum) error {
	return r.db.QueryRow(
		"INSERT INTO forums (slug, title, \"user\") VALUES ($1, $2, $3) RETURNING id",
		forum.Slug,
		forum.Title,
		forum.User,
	).Scan(&forum.ID)
}

func (r *Repository) CreateThread(thread *models.Thread) error {
	return r.db.QueryRow(
		"INSERT INTO threads (forum, author, created, message, title, slug, votes) " +
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		thread.Forum,
		thread.Author,
		thread.Created,
		thread.Message,
		thread.Title,
		thread.Slug,
		thread.Votes,
	).Scan(&thread.ID)
}

func (r *Repository) FindBySlug(slug string) (*models.Forum, error) {
	f := new(models.Forum)
	if err := r.db.QueryRow(
		"SELECT id, slug, title, \"user\", " +
			"(SELECT COUNT(*) FROM threads WHERE LOWER(forum) = LOWER($1)), " +
			"(SELECT COUNT(*) FROM posts WHERE LOWER(forum) = LOWER($1)) " +
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

func (r *Repository) FindByTitle(title string) (*models.Forum, error) {
	f := new(models.Forum)
	if err := r.db.QueryRow(
		"SELECT id, slug, title, \"user\" FROM forums WHERE LOWER(title) = LOWER($1)",
		title,
	).Scan(
		&f.ID,
		&f.Slug,
		&f.Title,
		&f.User,
	); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *Repository) GetThreads(slug string, params *models.ListParameters) ([]*models.Thread, error){
	var err error
	var rows *sql.Rows
	var threads []*models.Thread

	if params.Since != "" {
		layout := "2006-01-02T15:04:05Z07:00"
		t, err := time.Parse(layout, params.Since)
		if err != nil {
			return nil, err
		}

		if !params.Desc {
			rows, err = r.db.Query(
				"SELECT id, forum, author, created, message, title, slug, votes FROM threads "+
					"WHERE LOWER(forum) = LOWER($1) AND created >= $2 "+
					"ORDER BY "+
					"CASE WHEN $3 THEN created END DESC, "+
					"CASE WHEN NOT $3 THEN created END ASC "+
					"LIMIT CASE WHEN $4 > 0 THEN $4 END;",
				slug, t, params.Desc, params.Limit)
		} else {
			rows, err = r.db.Query(
				"SELECT id, forum, author, created, message, title, slug, votes FROM threads "+
					"WHERE LOWER(forum) = LOWER($1) AND created <= $2 "+
					"ORDER BY "+
					"CASE WHEN $3 THEN created END DESC, "+
					"CASE WHEN NOT $3 THEN created END ASC "+
					"LIMIT CASE WHEN $4 > 0 THEN $4 END;",
				slug, t, params.Desc, params.Limit)
		}

	} else {
		rows, err = r.db.Query(
			"SELECT id, forum, author, created, message, title, slug, votes FROM threads " +
				"WHERE LOWER(forum) = LOWER($1)" +
				"ORDER BY " +
				"CASE WHEN $2 THEN created END DESC, " +
				"CASE WHEN NOT $2 THEN created END ASC " +
				"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
			slug, params.Desc, params.Limit)
	}

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
	/*return r.db.QueryRow(
		`UPDATE threads SET votes = $1, title = $2, message = $3 
				WHERE id = $4 RETURNING id`,
		thread.Votes,
		thread.Title,
		thread.Message,
		thread.ID,
	).Scan(&thread.ID)*/
}

func (r * Repository) CreatePosts(posts []*models.Post) error {
	created := time.Now().UTC()
	for _, elem := range posts {
		elem.Created = created
		err := 	r.db.QueryRow(
			"INSERT INTO posts (path, author, created, forum, isEdited, message, parent, thread, slug) " +
				"VALUES (" +
				"CASE WHEN $6 > 0 THEN (SELECT P.path from posts AS P WHERE P.id = $6) || auto_id() ELSE auto_id() END, " +
				"$1, $2, $3, $4, $5, $6, $7, $8" +
				") RETURNING id",
			elem.Author,
			elem.Created,
			elem.Forum,
			elem.IsEdited,
			elem.Message,
			elem.Parent,
			elem.Thread,
			elem.Slug,
		).Scan(&elem.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) FindPost(id int64) (*models.Post, error) {
	p := new(models.Post)
	if err := r.db.QueryRow(
		"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts WHERE id = $1",
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
		&p.Slug,
	); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) FindPostBySlug(slug string) (*models.Post, error) {
	p := new(models.Post)
	if err := r.db.QueryRow(
		"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts " +
			"WHERE LOWER(slug) = LOWER($1)",
		slug,
	).Scan(
		&p.ID,
		&p.Author,
		&p.Created,
		&p.Forum,
		&p.IsEdited,
		&p.Message,
		&p.Parent,
		&p.Thread,
		&p.Slug,
	); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) UpdatePost(post *models.Post) error {
	return r.db.QueryRow(
		"UPDATE posts SET message = $1, isEdited = $2 WHERE id = $3 RETURNING id",
		post.Message,
		post.IsEdited,
		post.ID,
	).Scan(&post.ID)
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
	var err error
	var rows *sql.Rows
	var posts []*models.Post

	var t time.Time
	var sinceId int64
	var sinceIsDate bool

	if params.Since != "" {
		layout := "2006-01-02T15:04:05Z07:00"
		t, err = time.Parse(layout, params.Since)
		if err != nil {
			sinceId, err =  strconv.ParseInt(params.Since, 10, 64)
			if err != nil {
				return nil, err
			}
		} else {
			sinceIsDate = true
		}
	}

	if params.Sort == "tree" {
		rows, err = r.db.Query(
			`SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts 
						WHERE thread = $1 AND 
						      (($2 AND NOT $5 AND created >= $3) OR 
						       ($2 AND $5 AND created <= $3) OR 
						       (NOT $2 AND NOT $5 AND $4 > 0 AND path > (SELECT path FROM posts WHERE id = $4)) OR 
						       (NOT $2 AND $5 AND $4 > 0 AND path < (SELECT path FROM posts WHERE id = $4)) OR 
						       (NOT $2 AND $4 = 0))
						ORDER BY 
						         CASE WHEN NOT $5 THEN path END,
						         CASE WHEN $5 THEN path END DESC
						LIMIT $6;`,
			thread.ID, sinceIsDate, t, sinceId, params.Desc, params.Limit)
	} else if params.Sort == "parent_tree" {
		rows, err = r.db.Query(
			`SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts 
						WHERE substring(path,1,7) IN (
						      (SELECT path FROM posts WHERE parent = 0 AND thread = $1 AND (
						        ($2 AND NOT $5 AND created >= $3) OR 
						        ($2 AND $5 AND created <= $3) OR
						        (NOT $2 AND NOT $5 AND $4 > 0 AND path > (SELECT path FROM posts WHERE id = $4)) OR 
						    	(NOT $2 AND $5 AND $4 > 0 AND path < (SELECT substring(path,1,7) FROM posts WHERE id = $4)) OR
						       	(NOT $2 AND $4 = 0))
						       	ORDER BY
						      		CASE WHEN NOT $5 THEN path END,
						       		CASE WHEN $5 THEN path END DESC
						        LIMIT $6))
						ORDER BY 
						         CASE WHEN NOT $5 THEN path END,
						         CASE WHEN $5 THEN substring(path,1,7) END DESC, path`,
			thread.ID, sinceIsDate, t, sinceId, params.Desc, params.Limit)
	} else {
		rows, err = r.db.Query(
			`SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts 
						WHERE thread = $1 AND 
						      (($2 AND NOT $5 AND created >= $3) OR 
						       ($2 AND $5 AND created <= $3) OR 
						       (NOT $2 AND NOT $5 AND $4 > 0 AND id > $4) OR 
						       (NOT $2 AND $5 AND $4 > 0 AND id < $4) OR 
						       (NOT $2 AND $4 = 0))
						ORDER BY 
						         CASE WHEN NOT $5 THEN id END,
						         CASE WHEN $5 THEN id END DESC
						LIMIT $6;`,
			thread.ID, sinceIsDate, t, sinceId, params.Desc, params.Limit)
	}

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		p := new(models.Post)

		err := rows.Scan(&p.ID, &p.Author, &p.Created, &p.Forum, &p.IsEdited, &p.Message, &p.Parent, &p.Thread, &p.Slug)
		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *Repository) GetUsers(slug string, params models.ListParameters) ([]*models.User, error) {
	var err error
	var rows *sql.Rows
	var users []*models.User

	log.Println(params)

	rows, err = r.db.Query(
		`SELECT R.nickname, R.fullname, R.about, R.email FROM 
            	(
            	    SELECT U.nickname, U.fullname, U.about, U.email 
					FROM users AS U
					INNER JOIN (SELECT author FROM threads WHERE LOWER(forum) = $1) AS T
					ON LOWER(T.author) = LOWER(U.nickname)
					WHERE ($2 = '' OR (NOT $3 AND LOWER(U.nickname) > $2) OR ($3 AND LOWER(U.nickname) < $2))
					UNION
					SELECT U.nickname, U.fullname, U.about, U.email 
					FROM users AS U
					INNER JOIN (SELECT author FROM posts WHERE LOWER(forum) = $1) AS P
					ON LOWER(P.author) = LOWER(U.nickname)
					WHERE ($2 = '' OR (NOT $3 AND LOWER(U.nickname) > $2) OR ($3 AND LOWER(U.nickname) < $2))
            	) AS R				
				GROUP BY R.nickname, R.fullname, R.about, R.email
				ORDER BY 
				         CASE WHEN NOT $3 THEN LOWER(R.nickname) END,
				         CASE WHEN $3 THEN LOWER(R.nickname) END DESC
				LIMIT CASE WHEN $4 > 0 THEN $4 END;`,
		strings.ToLower(slug), strings.ToLower(params.Since), params.Desc, params.Limit)
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