CREATE table users (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(30) NOT NULL,
    email           VARCHAR(50) NOT NULL UNIQUE,
    passwordHash    VARCHAR(255) NOT NULL, -- BCRYPT output is 60 characters
    refreshToken    VARCHAR(255), -- nullable, no token until first login

    createdAt       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updatedAt       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)