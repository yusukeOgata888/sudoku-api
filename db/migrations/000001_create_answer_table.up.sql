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