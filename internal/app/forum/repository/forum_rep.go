package forum_rep

import (
	"database/sql"
	"github.com/efimovad/Forums.git/internal/app/forum"
	"github.com/efimovad/Forums.git/internal/models"
	"log"
	"strconv"
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
		"SELECT id, slug, title, \"user\" FROM forums WHERE LOWER(slug) = LOWER($1)",
		slug,
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
	return r.db.QueryRow(
		"UPDATE threads SET votes = $1, title = $2, message = $3 WHERE id = $4 RETURNING id",
		thread.Votes,
		thread.Title,
		thread.Message,
		thread.ID,
	).Scan(&thread.ID)
}

func Num2Str(num int64) string {
	if num == 0 {
		return ""
	}
	res := strconv.FormatInt(num, 10)
	for i := 0; i < MaxPostNum - len(res); i++ {
		res = "0" + res
	}
	return res
}

func (r * Repository) CreatePosts(posts []*models.Post) error {
	for _, elem := range posts {
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

func (r *Repository) CreateVote(vote *models.Vote) error {
	return r.db.QueryRow(
		"INSERT INTO votes (nickname, vote, thread) VALUES ($1, $2, $3) RETURNING id",
		vote.Nickname,
		vote.Voice,
		vote.Thread,
	).Scan(&vote.ID)
}

func (r *Repository) FindVote(thread string, nickname string) (*models.Vote, error) {
	v := new(models.Vote)
	if err := r.db.QueryRow(
		"SELECT id, nickname, vote, thread FROM votes " +
			"WHERE LOWER(nickname) = LOWER($1) AND LOWER(thread) = LOWER($2)",
		nickname, thread,
	).Scan(
		&v.ID,
		&v.Nickname,
		&v.Voice,
		&v.Thread,
	); err != nil {
		return nil, err
	}
	return v, nil
}

func (r *Repository) UpdateVote(vote *models.Vote) error {
	return r.db.QueryRow(
		"UPDATE votes SET vote = $1 WHERE id = $2 RETURNING id",
		vote.Voice,
		vote.ID,
	).Scan(&vote.ID)
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

	log.Println(params)
	if params.Sort == "tree" {
		rows, err = r.db.Query(
			`SELECT id, author, created, forum, isEdited, message, parent, thread, slug, path FROM posts 
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
		log.Println("$2:", sinceIsDate, "$3:", t, "$4:", sinceId, "$5:", params.Desc, "$6:", params.Limit)
		rows, err = r.db.Query(
			`SELECT id, author, created, forum, isEdited, message, parent, thread, slug, path FROM posts 
						WHERE thread = $1 AND 
						      (($2 AND NOT $5 AND created >= $3) OR 
						       ($2 AND $5 AND created <= $3) OR 
						       (NOT $2 AND NOT $5 AND $4 > 0 AND path > (SELECT path FROM posts WHERE id = $4) AND 
						        	substring(path from 1 for 6) in (select path from posts where parent = 0 and path > (SELECT path FROM posts WHERE id = $4) LIMIT $6)) OR 
						       (NOT $2 AND $5 AND $4 > 0 AND path < (SELECT path FROM posts WHERE id = $4)) OR 
						       (NOT $2 AND $4 = 0 AND path <= (SELECT MAX(b.path) || '999999'  from (SELECT path from posts where parent = 0 ORDER BY PATH LIMIT $6) b)::text ))
						ORDER BY 
						         CASE WHEN NOT $5 THEN path END,
						         CASE WHEN $5 THEN substring(path from 1 for 6) END DESC, path`,
			thread.ID, sinceIsDate, t, sinceId, params.Desc, params.Limit)
		/*if params.Since != "" {
			if !params.Desc {
				rows, err = r.db.Query(
					"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts "+
						"WHERE thread = $1 AND created >= $2"+
						"ORDER BY path ASC, id ASC " +
						"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
					thread.ID, t, params.Limit)
			} else {
				rows, err = r.db.Query(
					"WITH RECURSIVE T (author, created, forum, id, isEdited, message, parent, thread, PATHSTR, LEVEL, path ) " +
						"AS ( " +
						"SELECT T1.author, T1.created, T1.forum, T1.id, T1.isEdited, T1.message, T1.parent, T1.thread, " +
						"CAST (T1.path[1] AS VARCHAR (50)) as PATHSTR, 1, T1.path " +
						"FROM posts AS T1 WHERE T1.parent = 0 AND T1.thread = $1 AND created <= $2" +
						"UNION " +
						"SELECT  T2.author, T2.created, T2.forum, T2.id, T2.isEdited, T2.message, T2.parent, T2.thread, " +
						"CAST (T.PATHSTR || T2.path[T.LEVEL + 1] AS VARCHAR(50)), LEVEL + 1, T2.path " +
						"FROM posts AS T2 " +
						"INNER JOIN T " +
						"ON (T.id = T2.parent) " +
						") " +
						"SELECT author, created, forum, id, isEdited, message, parent, thread, slug " +
						"FROM T " +
						"ORDER BY T.path[1] ASC, T.PATHSTR DESC " +
						"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
					thread.ID, t, params.Limit)
			}

		} else {
			if params.Desc {
				rows, err = r.db.Query(
					"WITH RECURSIVE T (id, author, created, forum, isEdited, message, parent, thread, PATHSTR, LEVEL, path, slug ) "+
						"AS ( "+
						"SELECT T1.id, T1.author, T1.created, T1.forum, T1.isEdited, T1.message, T1.parent, T1.thread, "+
						"CAST (T1.path[1] AS VARCHAR (50)) as PATHSTR, 1, T1.path, T1.slug "+
						"FROM posts AS T1 WHERE T1.parent = 0 AND T1.thread = $1 "+
						"UNION "+
						"SELECT  T2.id, T2.author, T2.created, T2.forum, T2.isEdited, T2.message, T2.parent, T2.thread, "+
						"CAST (T.PATHSTR || T2.path[T.LEVEL + 1] AS VARCHAR(50)), LEVEL + 1, T2.path, T2.slug "+
						"FROM posts AS T2 "+
						"INNER JOIN T "+
						"ON (T.id = T2.parent) "+
						") "+
						"SELECT id, author, created, forum, isEdited, message, parent, thread, slug "+
						"FROM T "+
						"ORDER BY T.path[1] DESC, LENGTH(T.PATHSTR) ASC, T.PATHSTR ASC "+
						"LIMIT CASE WHEN $2 > 0 THEN $2 END;",
					thread.ID, params.Limit)
			} else {
				rows, err = r.db.Query(
					"WITH RECURSIVE T (id, author, created, forum, isEdited, message, parent, thread, PATHSTR, LEVEL, path, slug ) "+
						"AS ( "+
						"SELECT T1.id, T1.author, T1.created, T1.forum, T1.isEdited, T1.message, T1.parent, T1.thread, " +
						"CAST (T1.path[1] AS VARCHAR (50)) as PATHSTR, 1, T1.path, T1.slug " +
						"FROM posts AS T1 WHERE T1.parent = 0 AND T1.thread = $1 " +
						"UNION " +
						"SELECT T2.id, T2.author, T2.created, T2.forum, T2.isEdited, T2.message, T2.parent, T2.thread, " +
						"CAST (T.PATHSTR || T2.path[T.LEVEL + 1] AS VARCHAR(50)), LEVEL + 1, T2.path, T2.slug "+
						"FROM posts AS T2 " +
						"INNER JOIN T " +
						"ON (T.id = T2.parent) " +
						") " +
						"SELECT id, author, created, forum, isEdited, message, parent, thread, slug " +
						"FROM T " +
						"ORDER BY T.path[1] ASC, LENGTH(T.PATHSTR) ASC, T.PATHSTR ASC " +
						"LIMIT CASE WHEN $2 > 0 THEN $2 END;",
					thread.ID, params.Limit)
			}
		}*/
	} else {
		if params.Since != "" {
			log.Println(t)
			if !params.Desc {
				rows, err = r.db.Query(
					"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts "+
						"WHERE thread = $1 AND " +
						"($2 = 0 OR id > $2) "+
						"ORDER BY id ASC " +
						"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
					thread.ID, sinceId, params.Limit)
			} else {
				rows, err = r.db.Query(
					"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts "+
						"WHERE thread = $1 AND " +
						"($2 = 0 OR id < $2) "+
						"ORDER BY id DESC " +
						"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
					thread.ID, sinceId, params.Limit)
			}

		} else {
			rows, err = r.db.Query(
				"SELECT id, author, created, forum, isEdited, message, parent, thread, slug FROM posts "+
					"WHERE thread = $1"+
					"ORDER BY "+
					"CASE WHEN $2 THEN id END DESC, "+
					"CASE WHEN NOT $2 THEN id END ASC "+
					"LIMIT CASE WHEN $3 > 0 THEN $3 END;",
				thread.ID, params.Desc, params.Limit)
		}
	}

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		p := new(models.Post)

		if params.Sort == "tree" || params.Sort == "parent_tree" {
			var path string
			err := rows.Scan(&p.ID, &p.Author, &p.Created, &p.Forum, &p.IsEdited, &p.Message, &p.Parent, &p.Thread, &p.Slug, &path)
			if err != nil {
				return nil, err
			}
			log.Println(path)
		} else {
			err := rows.Scan(&p.ID, &p.Author, &p.Created, &p.Forum, &p.IsEdited, &p.Message, &p.Parent, &p.Thread, &p.Slug)
			if err != nil {
				return nil, err
			}
		}
		posts = append(posts, p)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return posts, nil
}