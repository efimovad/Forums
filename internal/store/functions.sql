DROP TRIGGER IF EXISTS on_vote_insert ON votes;
DROP TRIGGER IF EXISTS on_vote_update ON votes;


CREATE UNIQUE INDEX IF NOT EXISTS idx_votes_nickname_thread_unique
    ON votes (LOWER(nickname), thread);

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