CREATE table users (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(30) NOT NULL,
    email           VARCHAR(50) NOT NULL UNIQUE,
    passwordHash    VARCHAR(255) NOT NULL, -- BCRYPT output is 60 characters

    createdAt       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updatedAt       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    userID          INT NOT NULL,
    refreshToken    VARCHAR(255) NOT NULL UNIQUE,
    createdAt       DATETIME DEFAULT CURRENT_TIMESTAMP,
    expiresAt       DATETIME NOT NULL,

    FOREIGN KEY (userID) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_refreshToken (refreshToken),
    INDEX idx_userId (userID)
);