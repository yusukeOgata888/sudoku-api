CREATE TABLE `answers` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `cell_index` int NOT NULL,
  `cell_number` int NOT NULL,
  `session_id` varchar(255) NOT NULL
);

CREATE TABLE `problems` (
  `id` int NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `cell_index` int NOT NULL,
  `cell_number` int NOT NULL,
  `session_id` varchar(255) NOT NULL
);

CREATE TABLE submits (
  `session_id` varchar(36) NOT NULL,
  `user_id`    varchar(36) NOT NULL,
  `cell_index` tinyint     NOT NULL,
  `cell_number` tinyint    NOT NULL,
  PRIMARY KEY (`session_id`, `user_id`, `cell_index`)
);