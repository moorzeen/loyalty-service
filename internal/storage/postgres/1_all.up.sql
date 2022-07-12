create table USERS
(
    ID bigserial unique primary key,
    USERNAME text unique not null,
    PASSWORD_HASH bytea not null
);

create table SESSIONS
(
    USER_ID bigserial unique not null references USERS (ID),
    SIGN_KEY bytea not null
);

create table ORDERS
(
    ORDER_NUMBER text unique not null primary key,
    USER_ID bigserial not null references USERS (ID),
    UPLOADED_AT timestamptz not null default current_timestamp,
    STATUS text not null,
    ACCRUAL numeric default 0
);

create table ACCOUNTS
(
    USER_ID bigserial unique not null references USERS (ID),
    BALANCE numeric default 0,
    WITHDRAWN  numeric default 0
);

create table WITHDRAWALS
(
    USER_ID bigserial not null references USERS (ID),
    ORDER_NUMBER text unique not null,
    SUM numeric default 0,
    PROCESSED_AT timestamptz not null default current_timestamp
);