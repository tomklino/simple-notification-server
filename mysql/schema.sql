CREATE TABLE IF NOT EXISTS clients (
    client_id INT AUTO_INCREMENT,
    token VARCHAR(1023) NOT NULL,
    PRIMARY KEY (client_id)
)  ENGINE=INNODB;

CREATE TABLE IF NOT EXISTS topics (
    topic_id INT AUTO_INCREMENT,
    name VARCHAR(1023) NOT NULL,
    PRIMARY KEY (topic_id)
)  ENGINE=INNODB;

CREATE TABLE IF NOT EXISTS subsciptions (
    subsciption_id INT AUTO_INCREMENT,
    client_id INT NOT NULL,
    topic_id INT NOT NULL,
    PRIMARY KEY (subsciption_id)
)  ENGINE=INNODB;
