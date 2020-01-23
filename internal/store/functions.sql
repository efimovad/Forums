DROP TRIGGER IF EXISTS on_post_insert ON posts;
DROP TRIGGER IF EXISTS on_thread_insert ON threads;

CREATE TABLE IF NOT EXISTS forum_users (
    user_id BIGINT REFERENCES users(id),
    forum_id BIGINT REFERENCES forums(id)
);

DROP TRIGGER IF EXISTS on_vote_insert ON votes;
DROP TRIGGER IF EXISTS on_vote_update ON votes;

CREATE UNIQUE INDEX IF NOT EXISTS idx_forums_slug ON forums (LOWER(slug));
CREATE INDEX IF NOT EXISTS idx_forums_user ON forums ("user");

CREATE INDEX IF NOT EXISTS idx_threads_slug ON threads (LOWER(slug));
CREATE INDEX IF NOT EXISTS idx_threads_forum_created ON threads (LOWER(forum), created);
CREATE INDEX IF NOT EXISTS idx_threads_author ON threads (LOWER(author));

CREATE INDEX IF NOT EXISTS idx_users_nickname ON users (LOWER(nickname));

CREATE INDEX IF NOT EXISTS idx_posts_path ON posts USING GIN (path);
CREATE INDEX IF NOT EXISTS idx_posts_thread ON posts (thread);
CREATE INDEX IF NOT EXISTS idx_posts_forum ON posts (forum);
CREATE INDEX IF NOT EXISTS idx_posts_parent ON posts (parent);
CREATE INDEX IF NOT EXISTS idx_posts_thread_id ON posts (thread, id);

CREATE INDEX IF NOT EXISTS idx_votes_author ON votes (LOWER(nickname));
CREATE UNIQUE INDEX IF NOT EXISTS idx_votes_nickname_thread_unique ON votes (LOWER(nickname), thread);

CREATE OR REPLACE FUNCTION fn_update_thread_votes_ins()
    RETURNS TRIGGER AS '
    BEGIN
        UPDATE threads
        SET
            votes = votes + NEW.vote
        WHERE id = NEW.thread;
        RETURN NULL;
    END;
' LANGUAGE plpgsql;

CREATE TRIGGER on_vote_insert
    AFTER INSERT ON votes
    FOR EACH ROW EXECUTE PROCEDURE fn_update_thread_votes_ins();

CREATE OR REPLACE FUNCTION fn_update_thread_votes_upd()
    RETURNS TRIGGER AS '
    BEGIN
        IF OLD.vote = NEW.vote
        THEN
            RETURN NULL;
        END IF;
        UPDATE threads
        SET
            votes = votes + CASE WHEN NEW.vote = -1
                                     THEN -2
                                 ELSE 2 END
        WHERE id = NEW.thread;
        RETURN NULL;
    END;
' LANGUAGE plpgsql;

CREATE TRIGGER on_vote_update
    AFTER UPDATE ON votes
    FOR EACH ROW EXECUTE PROCEDURE fn_update_thread_votes_upd();

CREATE OR REPLACE FUNCTION forum_users_update()
    RETURNS TRIGGER AS '
    BEGIN
        INSERT INTO forum_users (user_id, forum_id) VALUES ((SELECT id FROM users WHERE LOWER(NEW.author) = LOWER(nickname)),
                                                        (SELECT id FROM forums WHERE LOWER(NEW.forum) = LOWER(slug)));
        RETURN NULL;
    END;
' LANGUAGE plpgsql;

CREATE TRIGGER on_post_insert
    AFTER INSERT ON posts
    FOR EACH ROW EXECUTE PROCEDURE forum_users_update();

CREATE TRIGGER on_thread_insert
    AFTER INSERT ON threads
    FOR EACH ROW EXECUTE PROCEDURE forum_users_update();