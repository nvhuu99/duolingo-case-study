CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    lastname VARCHAR(100) NOT NULL,
    firstname VARCHAR(100) NOT NULL,
    birthdate DATE NOT NULL,
    device_token VARCHAR(255) NOT NULL,
    native_language VARCHAR(20) NOT NULL,
    membership_id TINYINT NOT NULL,
    
    CONSTRAINT fk_membership FOREIGN KEY (membership_id) REFERENCES memberships(id)
);
