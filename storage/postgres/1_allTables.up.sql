create table USERS
(
    ID bigserial primary key,
    USER_LOGIN text unique not null,
    PASSWORD_HASH bytea not null
);

create table USER_SESSIONS
(
    USER_ID bigserial primary key,
    SIGN_KEY bytea not null
);

create table USER_ORDERS
(
    ID bigserial,
    ORDER_NUMBER bigint primary key,
    USER_LOGIN text unique not null references USERS (USER_LOGIN),
    UPLOADED_AT timestamptz not null default current_timestamp,
    STATUS text not null,
    ACCRUAL bigint not null default 0
);

create table BALANCES
(
    ID bigserial primary key,
    USER_ID bigint not null references USERS (ID),
    BALANCE bigint not null default 0,
    WITHDRAWN bigint not null default 0
);

create table WITHDRAWALS
(
    ID bigserial primary key,
    USER_ID bigint not null references USERS (ID),
    ORDER_NUMBER bigint references USER_ORDERS (ORDER_NUMBER),
    SUM bigint not null default 0,
    PROCESSED_AT timestamptz not null default current_timestamp
);