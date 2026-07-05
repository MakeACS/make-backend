-- +goose Up
-- +goose statementbegin
CREATE OR REPLACE FUNCTION create_default_makerspace_hours()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO default_hours (makerspace_id, day_of_week)
    VALUES
        (NEW.id, 0),
        (NEW.id, 1),
        (NEW.id, 2),
        (NEW.id, 3),
        (NEW.id, 4),
        (NEW.id, 5),
        (NEW.id, 6);
    
    RETURN NEW;
END;
$$ language plpgsql;
-- +goose statementend

CREATE TRIGGER after_makerspace_insert
AFTER INSERT on makerspaces
FOR EACH ROW EXECUTE FUNCTION create_default_makerspace_hours();


-- +goose Down
DROP TRIGGER IF EXISTS after_makerspace_insert ON makerspaces;
DROP FUNCTION IF EXISTS create_default_makerspace_hours();