CREATE TABLE IF NOT EXISTS "reservations" (
    id serial primary key,
    room_id int not null,
    start_time timestamp not null,
    end_time timestamp not null
);