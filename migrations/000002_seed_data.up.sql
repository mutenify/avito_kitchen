INSERT INTO restaurants (id, name, address, is_active) VALUES
(1, 'Додо Пицца', 'Невский проспект, 54', true),
(2, 'Бургер Кинг', 'Ефимова, 3', true);

-- Меню Додо Пиццы (id = 1)
INSERT INTO menu_items (restaurant_id, name, description, price, is_available) VALUES
(1, 'Пицца Пепперони', 'Пикантная пепперони, моцарелла, томатный соус', 599.00, true),
(1, 'Пицца Маргарита', 'Моцарелла, томаты, итальянские травы', 499.00, true),
(1, 'Кока-Кола 0.5', 'Освежающий напиток', 120.00, false); -- нет в наличии для теста корнер-кейса!

-- Меню Бургер Кинга (id = 2)
INSERT INTO menu_items (restaurant_id, name, description, price, is_available) VALUES
(2, 'Воппер', 'Легендарный бургер с говядиной на гриле', 349.00, true),
(2, 'Картофель Фри', 'Хрустящий картофель с солью', 119.00, true);
