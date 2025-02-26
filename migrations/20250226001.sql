DO
$$
BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'uint256') THEN
CREATE DOMAIN UINT256 AS NUMERIC
    CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
ELSE
ALTER DOMAIN UINT256 DROP CONSTRAINT uint256_check;
ALTER DOMAIN UINT256 ADD
    CHECK (VALUE >= 0 AND VALUE < POWER(CAST(2 AS NUMERIC), CAST(256 AS NUMERIC)) AND SCALE(VALUE) = 0);
END IF;
END
$$;


CREATE TABLE IF NOT EXISTS addresses
(
    guid         VARCHAR PRIMARY KEY,
    address      VARCHAR UNIQUE NOT NULL,
    address_type SMALLINT       NOT NULL DEFAULT 0,
    public_key   VARCHAR        NOT NULL,
    timestamp    INTEGER        NOT NULL CHECK (timestamp > 0)
    );
CREATE INDEX IF NOT EXISTS addresses_address ON addresses (address);
CREATE INDEX IF NOT EXISTS addresses_timestamp ON addresses (timestamp);
