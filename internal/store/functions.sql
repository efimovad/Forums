ALTER SEQUENCE post_path RESTART;

DROP TRIGGER IF EXISTS on_vote_insert ON votes;
DROP TRIGGER IF EXISTS on_vote_update ON votes;

CREATE UNIQUE INDEX IF NOT EXISTS idx_forums_slug ON forums (LOWER(slug));
CREATE INDEX IF NOT EXISTS idx_forums_user ON forums ("user");

CREATE INDEX IF NOT EXISTS idx_threads_slug ON threads (LOWER(slug));
CREATE INDEX IF NOT EXISTS idx_threads_forum_created ON threads (LOWER(forum), created);
CREATE INDEX IF NOT EXISTS idx_threads_author ON threads (LOWER(author));

CREATE INDEX IF NOT EXISTS idx_users_nickname ON users (LOWER(nickname));

CREATE INDEX IF NOT EXISTS idx_posts_path ON posts (path);
CREATE INDEX IF NOT EXISTS idx_posts_sub_path ON posts (substring(path,1,7));
CREATE INDEX IF NOT EXISTS idx_thread ON posts (thread);
CREATE INDEX IF NOT EXISTS idx_parent_thread ON posts (thread) WHERE parent = 0;

CREATE INDEX IF NOT EXISTS idx_posts_author ON posts (LOWER(author));
CREATE INDEX IF NOT EXISTS idx_posts_forum ON posts (LOWER(forum));

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