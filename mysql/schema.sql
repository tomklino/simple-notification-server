CREATE TABLE IF NOT EXISTS clients (
    client_id INT AUTO_INCREMENT,
    token VARCHAR(1023) UNIQUE NOT NULL,
    PRIMARY KEY (client_id)
)  ENGINE=INNODB;

CREATE TABLE IF NOT EXISTS topics (
    topic_id INT AUTO_INCREMENT,
    name VARCHAR(1023) UNIQUE NOT NULL,
    PRIMARY KEY (topic_id)
)  ENGINE=INNODB;

CREATE TABLE IF NOT EXISTS subscriptions (
    subsciption_id INT AUTO_INCREMENT,
    client_id INT NOT NULL,
    FOREIGN KEY (client_id)
        REFERENCES clients(client_id)
        ON UPDATE CASCADE,
    topic_id INT NOT NULL,
    FOREIGN KEY (topic_id)
        REFERENCES topics(topic_id)
        ON UPDATE CASCADE,
    PRIMARY KEY (subsciption_id),
    UNIQUE (client_id, topic_id)
)  ENGINE=INNODB;
