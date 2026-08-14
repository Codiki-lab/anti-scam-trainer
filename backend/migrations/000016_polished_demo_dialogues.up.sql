CREATE TABLE migration_000016_archived_chats AS
SELECT id,content_status,archived_at,is_active
FROM chats
WHERE content_status='published' AND archived_at IS NULL
  AND product_context->>'content_version'='issue-103-complete';
ALTER TABLE migration_000016_archived_chats ADD PRIMARY KEY(id);

UPDATE chats
SET content_status='archived',archived_at=CURRENT_TIMESTAMP,is_active=FALSE
WHERE id IN (SELECT id FROM migration_000016_archived_chats);

CREATE TABLE migration_000016_new_chats (id INTEGER PRIMARY KEY REFERENCES chats(id) ON DELETE CASCADE);

WITH inserted AS (
    INSERT INTO chats(title,description,difficulty,role,is_active,level_id,user_role,content_status,scam_scheme,product_context,ai_system_prompt,final_rubric,topic_id,risk_type)
    SELECT c.title,c.description,c.difficulty,c.role,TRUE,c.level_id,c.user_role,'published',c.scam_scheme,c.product_context,c.ai_system_prompt,c.final_rubric,c.topic_id,c.risk_type
    FROM chats c JOIN migration_000016_archived_chats m ON m.id=c.id
    RETURNING id
)
INSERT INTO migration_000016_new_chats SELECT id FROM inserted;

INSERT INTO chat_steps(chat_id,step_number,response_type,step_goal,counterparty_message,ai_instruction,fallback_message,max_points)
SELECT replacement.id,st.step_number,st.response_type,st.step_goal,
       CASE WHEN t.slug='seller-fake-payment' AND st.step_number=2
            THEN 'Откройте форму получения 67 000 ₽ и укажите карту, на которую зачислить деньги.'
            ELSE st.counterparty_message END,
       st.ai_instruction,st.fallback_message,st.max_points
FROM chat_steps st
JOIN chats original ON original.id=st.chat_id
JOIN migration_000016_archived_chats archived ON archived.id=original.id
JOIN topics t ON t.id=original.topic_id
JOIN chats replacement ON replacement.topic_id=original.topic_id AND replacement.level_id=original.level_id
JOIN migration_000016_new_chats created ON created.id=replacement.id;

CREATE TABLE migration_000016_reactions (
    topic_slug VARCHAR PRIMARY KEY,
    opening_safe TEXT NOT NULL,
    opening_cautious TEXT NOT NULL,
    opening_unsafe TEXT NOT NULL,
    request_safe TEXT NOT NULL,
    request_cautious TEXT NOT NULL,
    request_unsafe TEXT NOT NULL
);

INSERT INTO migration_000016_reactions VALUES
('buyer-phishing-links','Через приложение получится дольше. Сейчас объясню, как открыть готовую форму.','Телефон исправен, но для доставки всё равно понадобится отдельное подтверждение.','Отлично, сейчас отправлю форму — так оплатить быстрее.','Без моей формы заказ не появится. Решайте быстрее.','Пока вы проверяете, другой покупатель уже готов оплатить.','Подтвердите оплату сейчас, пока форма ещё доступна.'),
('buyer-prepayment','Без перевода бронь не поставлю: желающих на приставку много.','Комплект полный, но решение нужно принять сегодня.','Хорошо, сейчас пришлю реквизиты для брони.','Пока вы проверяете, я не могу обещать, что товар останется.','Проверка займёт время, а следующий покупатель уже ждёт реквизиты.','Переводите сейчас — после этого сразу сниму объявление.'),
('buyer-fake-delivery','Обычная доставка сейчас недоступна, поэтому нужна отдельная страховка.','Комплект полный. Курьер сможет приехать после оплаты страховки.','Отлично, сейчас пришлю форму курьерской службы.','В поддержке будут отвечать долго, а курьер уже ждёт.','Название услуги есть в форме, откройте её и посмотрите.','Оплатите сейчас, иначе курьер уедет и заказ отменится.'),
('buyer-off-platform','В чате видео не загружается. Проще перейти в Telegram.','Камера исправна, полный комплект покажу на видео в Telegram.','Хорошо, сейчас пришлю контакт и QR для просмотра.','Без Telegram и QR видео посмотреть не получится.','QR ведёт только на видео, проверять там нечего.','Переходите сейчас: скидка на доставку скоро исчезнет.'),
('buyer-sms-codes','Самостоятельно отменить второй заказ не получится: нужен код из сообщения.','Лишний заказ уже создан, поэтому важно действовать быстро.','Хорошо, сейчас запрошу код для отмены.','Пока вы проверяете, деньги могут списаться повторно.','Назначение кода неважно, он нужен только для отмены.','Продиктуйте код сейчас, иначе отмена не успеет пройти.'),
('buyer-too-good-offer','Штатно оформить не получится: из-за срочного отъезда нужна бронь через форму.','Цена такая только сегодня, долго проверять предложение не получится.','Отлично, внесите бронь — и ноутбук останется за вами.','Если не открыть форму сейчас, курьер заберёт другой заказ.','Проверяйте быстрее: другой покупатель уже готов внести деньги.','Оплатите сейчас, пока низкая цена ещё действует.'),
('seller-fake-payment','Я уже всё оплатила и отправила курьера. Проверьте сообщение о получении денег.','Курьер приедет сегодня, а оплату я уже отправила.','Отлично, тогда курьер сможет сразу забрать телефон.','Но деньги уже списаны; без подтверждения они зависнут.','Форма обычная, данные карты нужны только для зачисления.','Введите реквизиты, затем останется подтвердить получение денег.'),
('seller-fake-delivery','Заказ уже оформлен, но продавцу нужно отдельно подтвердить страховку курьера.','Курьер приедет сегодня, точное время появится после подтверждения.','Отлично, тогда курьер заберёт кофемашину без задержки.','В заказе страховка не отображается, потому что она относится к курьерской службе.','Поддержка ответит не сразу, а курьер уже назначен.','Оплатите страховку сейчас, после доставки деньги вернутся.'),
('seller-external-links','В разделе продаж заказ появится только после активации по моей странице.','Курьер назначен, а статус обновится после подтверждения оплаты.','Хорошо, сейчас пришлю страницу с подтверждением заказа.','Без страницы деньги не зачислятся на ваш счёт.','На странице нужно только подтвердить карту получателя.','Активируйте оплату сейчас, иначе она зависнет.'),
('seller-confirmation-codes','Самостоятельная проверка займёт время, а профиль могут заблокировать раньше.','Вижу ошибку подтверждения; специалист уже готов помочь.','Хорошо, сейчас специалист запросит код для подтверждения.','Официальная поддержка ответит позже, а заказ отменится через несколько минут.','Код нужен только для проверки статуса профиля.','Перешлите код сейчас, чтобы заказ не отменился.'),
('seller-off-platform','В этом чате проблема не решится: специалист отвечает только в WhatsApp.','Видео снова не откроется, лучше сразу перейти в другой мессенджер.','Отлично, сейчас пришлю номер специалиста в WhatsApp.','Без номера специалист не сможет найти ваш заказ.','Это служебный контакт, общение там полностью безопасно.','Переходите сейчас, пока обращение не закрыли.'),
('seller-pressure','Я уже перевёл деньги и отправил курьера; ждать проверки он не сможет.','Курьер будет через несколько минут, перевод уже должен прийти.','Отлично, тогда курьер сразу заберёт камеру.','Пока вы открываете банк, за ожидание курьера уже начисляется плата.','Скриншот подтверждает перевод, дополнительная проверка не нужна.','Отдайте камеру сейчас, иначе с меня спишут комиссию.');

INSERT INTO chat_options(step_id,option_text,counterparty_reaction,explanation,points,sort_order)
SELECT replacement_step.id,
       CASE
         WHEN original_option.option_text ~ '^Точно: ' THEN regexp_replace(original_option.option_text,'^Точно: ','')||CASE original_option.sort_order
           WHEN 1 THEN ' Других вариантов не рассматриваю.'
           WHEN 2 THEN ' До проверки ничего предпринимать не буду.'
           ELSE ' Готов действовать без задержки.' END
         WHEN original_option.option_text ~ '^Сразу скажу: ' THEN regexp_replace(original_option.option_text,'^Сразу скажу: ','')||CASE original_option.sort_order
           WHEN 1 THEN ' Продолжим только на этих условиях.'
           WHEN 2 THEN ' После этого решу, продолжать ли сделку.'
           ELSE ' Можно действовать сразу.' END
         ELSE original_option.option_text
       END,
       CASE WHEN original_option.counterparty_reaction IS NULL THEN NULL
            WHEN original_step.step_number=1 THEN (ARRAY[r.opening_safe,r.opening_cautious,r.opening_unsafe])[original_option.sort_order]
            ELSE (ARRAY[r.request_safe,r.request_cautious,r.request_unsafe])[original_option.sort_order] END,
       original_option.explanation,original_option.points,original_option.sort_order
FROM chat_options original_option
JOIN chat_steps original_step ON original_step.id=original_option.step_id
JOIN chats original ON original.id=original_step.chat_id
JOIN migration_000016_archived_chats archived ON archived.id=original.id
JOIN topics t ON t.id=original.topic_id
JOIN migration_000016_reactions r ON r.topic_slug=t.slug
JOIN chats replacement ON replacement.topic_id=original.topic_id AND replacement.level_id=original.level_id
JOIN migration_000016_new_chats created ON created.id=replacement.id
JOIN chat_steps replacement_step ON replacement_step.chat_id=replacement.id AND replacement_step.step_number=original_step.step_number;

DROP TABLE migration_000016_reactions;
