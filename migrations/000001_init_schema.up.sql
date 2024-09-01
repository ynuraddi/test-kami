CREATE TABLE IF NOT EXISTS "reservations" (
    id serial primary key,
    room_id int not null,
    start_time timestamp not null,
    end_time timestamp not null
);

CREATE INDEX idx_reservations_room ON reservations (room_id);

CREATE INDEX idx_reservations_room_start_end ON reservations (room_id, start_time, end_time);