CREATE TABLE migration_000010_content_seed (
    topic_slug VARCHAR PRIMARY KEY,
    title VARCHAR NOT NULL,
    description TEXT NOT NULL,
    risk_type VARCHAR NOT NULL,
    item TEXT NOT NULL,
    price INTEGER NOT NULL,
    counterpart TEXT NOT NULL,
    opening TEXT NOT NULL,
    risk_signal TEXT NOT NULL,
    deceptive_request TEXT NOT NULL,
    pressure TEXT NOT NULL,
    safe_action TEXT NOT NULL,
    verification TEXT NOT NULL,
    cautious_action TEXT NOT NULL,
    unsafe_action TEXT NOT NULL,
    summary TEXT NOT NULL
);

INSERT INTO migration_000010_content_seed VALUES
('buyer-phishing-links','Фишинговые ссылки','Как отличить поддельную страницу оплаты или доставки от штатного экрана Avito.','phishing','iPhone 14, 128 ГБ',42000,'Максим','Телефон в наличии. Могу отправить Avito Доставкой уже сегодня.','Собеседник присылает внешний адрес, похожий на Avito, а заказа внутри приложения нет.','Откройте присланную форму и оплатите заказ банковской картой.','Ссылка действует несколько минут, потом телефон заберёт другой покупатель.','Не открывать ссылку и проверить заказ из карточки объявления внутри приложения.','Самостоятельно открыть раздел заказов и убедиться, что покупка действительно создана.','Сверить профиль и описание, но пока продолжить переписку.','Перейти по ссылке и ввести данные карты.','Заказ, оплата и доставка проверяются только внутри самостоятельно открытого приложения.'),
('buyer-prepayment','Предоплата','Как не потерять деньги на брони товара до проверки продавца и штатного оформления.','prepayment','PlayStation 5 Slim',39000,'Артём','Приставка новая, желающих много, могу придержать её для вас.','Незнакомый продавец требует перевод за бронь до осмотра или заказа в приложении.','Переведите 3 000 ₽ на карту, и я сниму объявление.','Решайте сейчас: следующий покупатель уже просит реквизиты.','Не переводить бронь; предложить штатное оформление или оплату после проверки при встрече.','Проверить возможность безопасной покупки в приложении и осмотреть товар до оплаты.','Попросить дополнительные фото и документы, не переводя деньги сразу.','Перевести предоплату продавцу напрямую.','Предоплата незнакомому человеку не гарантирует товар; безопаснее штатное оформление или проверка при встрече.'),
('buyer-fake-delivery','Поддельная доставка','Как распознать выдуманную курьерскую процедуру и неожиданный платёж покупателя.','delivery','Пылесос Dyson V12',33000,'Ирина','Пылесос упакован, доставку можно оформить сегодня.','Собеседник выдумывает страховку, активацию или комиссию, которой нет в заказе Avito.','Оплатите возвратную страховку курьера по моей форме.','Курьер уже ждёт; без оплаты страховки заказ отменится.','Не оплачивать стороннюю услугу и открыть условия доставки внутри приложения.','Проверить заказ, стоимость и услуги доставки самостоятельно в приложении.','Уточнить название услуги в официальной поддержке, не открывая форму.','Оплатить обещанную возвратную страховку.','Покупатель не оплачивает комиссии и страховки по инструкции собеседника.'),
('buyer-off-platform','Общение вне сервиса','Почему перевод сделки в Telegram или WhatsApp повышает риск подмены контакта и оформления.','external_messenger','Фотоаппарат Canon EOS R10',68000,'Ольга','Фотоаппарат в отличном состоянии, могу прислать подробное видео.','Собеседник уводит переписку во внешний мессенджер и продолжает оформление там.','Напишите в Telegram, там пришлю видео и QR для покупки.','Скидка на доставку доступна только сегодня и только по QR.','Остаться в чате Avito и попросить отправить материалы и оформить сделку здесь.','Проверить заказ из карточки объявления, не сканируя QR и не переходя в мессенджер.','Уточнить детали товара в текущем чате без передачи контакта.','Перейти в мессенджер и отсканировать QR.','Переписка и оформление внутри сервиса сохраняют проверяемую историю сделки.'),
('buyer-sms-codes','SMS-коды','Как отличить код входа или подтверждения от выдуманного кода отмены заказа.','account_takeover','AirPods Pro 2',17000,'Денис','Наушники доступны, но приложение якобы создало два заказа.','Собеседник просит код из SMS, хотя сообщение предупреждает, что его нельзя передавать.','Назовите шестизначный код, чтобы я отменил лишний заказ.','Без кода деньги спишутся повторно через несколько минут.','Не сообщать код, самостоятельно проверить заказы и обратиться в поддержку из приложения.','Прочитать назначение кода в SMS и открыть раздел заказов самостоятельно.','Спросить, какое действие подтверждает код, не называя его.','Передать код собеседнику для отмены заказа.','Любой код подтверждает действие владельца аккаунта и не нужен продавцу, курьеру или поддержке в чате.'),
('buyer-too-good-offer','Слишком выгодное предложение','Как заметить сочетание нереальной цены, срочности и нештатной оплаты.','social_engineering','MacBook Air M3',55000,'Кирилл','Ноутбук почти новый, цена ниже рынка из-за срочного отъезда.','Сильная скидка сочетается с дефицитом времени, предоплатой и отказом от штатной проверки.','Внесите 5 000 ₽ через внешнюю форму, чтобы курьер забрал ноутбук.','Через три минуты предложение получит другой покупатель.','Взять паузу, отказаться от ссылки и предоплаты и предложить штатную сделку.','Сравнить цену, проверить профиль и наличие заказа внутри приложения.','Задать вопросы о товаре и попросить время на проверку.','Заплатить за бронь, чтобы не потерять выгодную цену.','Выгода и срочность не меняют безопасный порядок проверки и оплаты.'),
('seller-fake-payment','Поддельная оплата','Как отличить реальное поступление денег от скриншота, SMS или страницы покупателя.','fake_payment','iPhone 15, 128 ГБ',67000,'Анна','Телефон беру без торга, оплату отправлю прямо сейчас.','Покупатель показывает чек или ссылку, но денег в самостоятельно открытом банке или заказе нет.','Откройте форму получения 67 000 ₽ и подтвердите свою карту.','Курьер уже едет; если не подтвердите, с меня спишут комиссию.','Не открывать форму и не отдавать товар до фактического подтверждения оплаты.','Самостоятельно проверить заказ в приложении и поступление в своём банке.','Сообщить, что товар будет передан только после проверки денег.','Довериться чеку и передать товар курьеру.','Скриншот, SMS и ссылка покупателя не подтверждают получение денег.'),
('seller-fake-delivery','Фальшивая доставка','Как продавцу распознать выдуманную страховку курьера и платёж за передачу товара.','delivery','Кофемашина De’Longhi',31000,'Михаил','Кофемашину заберёт курьер, доставку я уже оформил.','Покупатель требует от продавца неожиданный возвратный платёж, отсутствующий в заказе.','Оплатите 2 490 ₽ страховки, после этого курьер увидит адрес.','Форма действует пять минут, затем курьер уедет.','Не платить и проверить условия и заказ внутри приложения.','Самостоятельно открыть заказ и официальные правила доставки.','Уточнить условия в поддержке, открытой из приложения.','Оплатить страховку, которую обещают вернуть.','Продавец не оплачивает покупателю страховку, комиссию или активацию доставки.'),
('seller-external-links','Внешние ссылки','Как не открыть поддельную форму получения денег и не раскрыть реквизиты карты.','phishing','Велосипед Trek Marlin 6',54000,'Павел','Велосипед подходит, я оформил доставку и оплатил заказ.','Заказа в приложении нет, а покупатель присылает внешнюю форму получения денег.','Откройте страницу и введите номер карты, срок и CVC для зачисления.','Оплата зависнет, если не активировать её за три минуты.','Не открывать ссылку, не вводить реквизиты и проверить заказ самостоятельно.','Открыть раздел продаж в приложении и убедиться, что заказ и оплата существуют.','Попросить покупателя оформить заказ повторно штатным способом.','Ввести реквизиты карты на странице покупателя.','Для получения денег продавцу не нужны CVC, код из SMS или внешняя форма.'),
('seller-confirmation-codes','Коды подтверждения','Почему покупателю, курьеру и поддержке не нужен код из SMS продавца.','account_takeover','PlayStation 5',44000,'Сергей','Приставку беру, но профиль продавца якобы требует подтверждения.','Внешний контакт под видом поддержки просит код входа или подтверждения операции.','Перешлите код специалисту, чтобы покупатель смог завершить оплату.','Без кода профиль заблокируют и заказ отменят.','Никому не сообщать код и открыть поддержку самостоятельно внутри приложения.','Прочитать назначение кода и проверить активные сеансы и заказы в аккаунте.','Прервать разговор и уточнить статус в официальной поддержке.','Отправить код покупателю или специалисту из сообщения.','Код предназначен только владельцу аккаунта и подтверждает действие от его имени.'),
('seller-off-platform','Общение вне сервиса','Как сохранить историю сделки и не попасть к лжеподдержке во внешнем мессенджере.','external_messenger','Ноутбук ASUS VivoBook',48000,'Елена','Ноутбук подходит, но видео в чате у меня не открывается.','Покупатель просит номер телефона, переводит продавца в WhatsApp и подключает лжеподдержку.','Напишите в WhatsApp; специалист там подтвердит ваш профиль.','Поддержка уже ждёт и закроет обращение через несколько минут.','Не передавать контакт и продолжить общение и оформление внутри Avito.','Самостоятельно открыть поддержку и проверить заказ в приложении.','Попросить описать проблему в текущем чате без передачи номера.','Перейти в WhatsApp и следовать указаниям контакта поддержки.','Официальная поддержка открывается самостоятельно, а история сделки остаётся в чате сервиса.'),
('seller-pressure','Давление','Как остановить передачу товара, когда покупатель торопит курьером, чеком или угрозой комиссии.','social_engineering','Sony Alpha A7 III',82000,'Роман','Камеру беру, перевод уже якобы отправлен, курьер будет скоро.','Срочность и чужие финансовые последствия заменяют самостоятельную проверку сделки.','Передайте камеру курьеру по скриншоту перевода.','Курьер у подъезда; за задержку с покупателя спишут деньги.','Взять паузу и не отдавать товар до фактической проверки оплаты и заказа.','Самостоятельно открыть банк и раздел продаж, не используя материалы покупателя.','Сообщить курьеру, что передача возможна только после проверки.','Отдать товар, чтобы избежать чужой комиссии.','Таймер, курьер и давление не подтверждают оплату и не меняют безопасный порядок передачи товара.');

CREATE TABLE migration_000010_topic_snapshot AS
SELECT id,title,description FROM topics WHERE slug IN (SELECT topic_slug FROM migration_000010_content_seed);
ALTER TABLE migration_000010_topic_snapshot ADD PRIMARY KEY(id);

CREATE TABLE migration_000010_theory_snapshot AS
SELECT id,kind,title,body FROM theory_blocks WHERE topic_id IN (SELECT id FROM topics WHERE slug IN (SELECT topic_slug FROM migration_000010_content_seed));
ALTER TABLE migration_000010_theory_snapshot ADD PRIMARY KEY(id);

CREATE TABLE migration_000010_question_snapshot AS
SELECT id,text,explanation FROM quiz_questions WHERE topic_id IN (SELECT id FROM topics WHERE slug IN (SELECT topic_slug FROM migration_000010_content_seed));
ALTER TABLE migration_000010_question_snapshot ADD PRIMARY KEY(id);

CREATE TABLE migration_000010_quiz_option_snapshot AS
SELECT id,text,is_correct FROM quiz_options WHERE question_id IN (SELECT id FROM migration_000010_question_snapshot);
ALTER TABLE migration_000010_quiz_option_snapshot ADD PRIMARY KEY(id);

CREATE TABLE migration_000010_free_play_snapshot AS SELECT * FROM free_play_configs;
ALTER TABLE migration_000010_free_play_snapshot ADD PRIMARY KEY(user_role);

UPDATE topics t SET title=s.title,description=s.description
FROM migration_000010_content_seed s WHERE t.slug=s.topic_slug;

UPDATE theory_blocks b SET
    kind=(ARRAY['intro','risk','example','safe_action','summary'])[b.sort_order],
    title=(ARRAY['Как устроена схема','Сигналы риска','Ситуация на Avito','Безопасное действие','Памятка'])[b.sort_order],
    body=CASE b.sort_order
        WHEN 1 THEN s.description||' В учебной ситуации объявление предлагает '||s.item||' за '||s.price||' ₽, а собеседником выступает '||s.counterpart||'. Задача Пользователя — отличить обычное обсуждение от нештатного оформления.'
        WHEN 2 THEN s.risk_signal||' Сам по себе профиль, вежливый тон или фотографии товара не подтверждают просьбу собеседника. Решение принимается по фактическому статусу сделки.'
        WHEN 3 THEN s.opening||' '||s.deceptive_request||' '||s.pressure
        WHEN 4 THEN s.safe_action||' '||s.verification
        ELSE s.summary||' При сомнении опасное действие останавливают до проверки, а не исправляют после передачи денег, товара или секрета.' END
FROM topics t JOIN migration_000010_content_seed s ON s.topic_slug=t.slug
WHERE b.topic_id=t.id;

UPDATE quiz_questions q SET
    text=CASE q.sort_order
        WHEN 1 THEN 'Какой главный признак риска в теме «'||s.title||'» для '||CASE t.user_role WHEN 'buyer' THEN 'покупателя' ELSE 'продавца' END||'?'
        WHEN 2 THEN 'Какое действие наиболее полно устраняет риск в теме «'||s.title||'» для '||CASE t.user_role WHEN 'buyer' THEN 'покупателя' ELSE 'продавца' END||'?'
        WHEN 3 THEN 'Как самостоятельно проверить сделку по теме «'||s.title||'» для '||CASE t.user_role WHEN 'buyer' THEN 'покупателя' ELSE 'продавца' END||'?'
        WHEN 4 THEN 'Какое действие в теме «'||s.title||'» для '||CASE t.user_role WHEN 'buyer' THEN 'покупателя' ELSE 'продавца' END||' опасно?'
        ELSE 'Какое правило нужно запомнить по теме «'||s.title||'» для '||CASE t.user_role WHEN 'buyer' THEN 'покупателя' ELSE 'продавца' END||'?' END,
    explanation=CASE q.sort_order
        WHEN 1 THEN s.risk_signal WHEN 2 THEN s.safe_action WHEN 3 THEN s.verification
        WHEN 4 THEN 'Опасное действие: '||s.unsafe_action ELSE s.summary END
FROM topics t JOIN migration_000010_content_seed s ON s.topic_slug=t.slug
WHERE q.topic_id=t.id;

UPDATE quiz_options o SET
    text=CASE WHEN o.sort_order=v.correct_order THEN v.correct_text
        ELSE v.wrong_options[CASE WHEN o.sort_order<v.correct_order THEN o.sort_order ELSE o.sort_order-1 END] END,
    is_correct=o.sort_order=v.correct_order
FROM quiz_questions q
JOIN topics t ON t.id=q.topic_id
JOIN migration_000010_content_seed s ON s.topic_slug=t.slug
CROSS JOIN LATERAL (SELECT
    ((t.sort_order+q.sort_order-2)%4)+1 AS correct_order,
    CASE q.sort_order WHEN 1 THEN s.risk_signal WHEN 2 THEN s.safe_action WHEN 3 THEN s.verification WHEN 4 THEN s.unsafe_action ELSE s.summary END AS correct_text,
    CASE q.sort_order
      WHEN 1 THEN ARRAY['Цена и фотографии «'||s.item||'» сами по себе','Вежливый тон собеседника по имени '||s.counterpart,'Само наличие объявления о товаре «'||s.item||'»']
      WHEN 2 THEN ARRAY[s.deceptive_request,s.cautious_action,s.unsafe_action]
      WHEN 3 THEN ARRAY['Попросить '||s.counterpart||' прислать скриншот','Следовать инструкции собеседника: '||s.deceptive_request,s.cautious_action]
      WHEN 4 THEN ARRAY[s.safe_action,s.verification,s.cautious_action]
      ELSE ARRAY['Для темы «'||s.title||'» срочность важнее проверки','Инструкцию собеседника можно считать штатным оформлением','Можно выполнить часть опасной просьбы и остановиться позже'] END AS wrong_options
) v
WHERE o.question_id=q.id;

UPDATE free_play_configs SET
    product_context=CASE user_role
      WHEN 'buyer' THEN '{"content_version":"issue-103-complete","item":"товар из объявления","channel":"чат Avito","situations":["штатное оформление","предоплата","внешняя ссылка","обычная сделка"]}'::jsonb
      ELSE '{"content_version":"issue-103-complete","item":"товар пользователя","channel":"чат Avito","situations":["штатная оплата","ложный чек","код подтверждения","обычная сделка"]}'::jsonb END,
    system_prompt='Веди короткий реалистичный диалог о сделке на Avito. Тип собеседника установлен сервером; не раскрывай учебную механику.',
    final_rubric=CASE user_role
      WHEN 'buyer' THEN '{"safe_channel":100,"checks_order":100,"blind_refusal":50,"external_payment":0,"shares_secret":0}'::jsonb
      ELSE '{"checks_payment":100,"keeps_item":100,"blind_refusal":50,"trusts_screenshot":0,"shares_secret":0}'::jsonb END;

CREATE TABLE migration_000010_archived_chats (
    id INTEGER PRIMARY KEY REFERENCES chats(id) ON DELETE CASCADE,
    previous_content_status VARCHAR NOT NULL,
    previous_archived_at TIMESTAMP WITH TIME ZONE,
    previous_is_active BOOLEAN NOT NULL
);

INSERT INTO migration_000010_archived_chats
SELECT id,content_status,archived_at,is_active FROM chats
WHERE content_status='published' AND archived_at IS NULL AND (
    product_context?'content_key' OR product_context->>'seed_version'='issue-49' OR
    (title='Покупатель: проверка оплаты' AND product_context='{"category":"электроника"}'::jsonb) OR
    (title='Покупатель: смешанный ответ' AND product_context='{"item":"смартфон","price":42000}'::jsonb) OR
    (title='Покупатель: свободный диалог' AND product_context='{"item":"игровая приставка","price":65000}'::jsonb) OR
    (title='Продавец: смешанный ответ' AND product_context='{"item":"кофемашина","price":18000}'::jsonb) OR
    (title='Продавец: свободный диалог' AND product_context='{"item":"велосипед","price":35000}'::jsonb)
);

UPDATE chats SET content_status='archived',archived_at=CURRENT_TIMESTAMP,is_active=FALSE
WHERE id IN (SELECT id FROM migration_000010_archived_chats);

CREATE TABLE migration_000010_new_chats (id INTEGER PRIMARY KEY REFERENCES chats(id) ON DELETE CASCADE);

WITH inserted AS (
    INSERT INTO chats(title,description,difficulty,role,is_active,level_id,user_role,content_status,scam_scheme,product_context,ai_system_prompt,final_rubric,topic_id)
    SELECT s.title||': '||(ARRAY['основы','выбор без подсказки','свободный ответ','диалог с мошенником'])[l.level_number],
           s.description,l.level_number::text,t.user_role,TRUE,l.id,t.user_role,'published',s.risk_type,
           jsonb_build_object('content_version','issue-103-complete','topic',s.topic_slug,'item',s.item,'price',s.price,'counterparty',s.counterpart,'channel','чат Avito'),
           'Используй только policy ролевой ветки, факты текущего Сценария и разрешённые сервером реплики.',
           jsonb_build_object('risk_type',s.risk_type,'risk_signal',s.risk_signal,'safe_action',s.safe_action),t.id
    FROM migration_000010_content_seed s JOIN topics t ON t.slug=s.topic_slug CROSS JOIN levels l
    WHERE NOT EXISTS (SELECT 1 FROM chats existing WHERE existing.topic_id=t.id AND existing.level_id=l.id AND existing.content_status='published' AND existing.archived_at IS NULL)
    RETURNING id
)
INSERT INTO migration_000010_new_chats SELECT id FROM inserted;

INSERT INTO chat_steps(chat_id,step_number,response_type,step_goal,counterparty_message,ai_instruction,fallback_message,max_points)
SELECT c.id,n,
       CASE l.level_number WHEN 1 THEN 'multiple_choice' WHEN 2 THEN 'similar_choice' WHEN 3 THEN CASE WHEN n=1 THEN 'multiple_choice' ELSE 'free_text' END ELSE 'free_text' END,
       CASE l.level_number
         WHEN 1 THEN 'Распознать риск и выбрать безопасное действие на этапе '||n
         WHEN 2 THEN 'Сравнить похожие решения и выбрать наиболее безопасное на этапе '||n
         WHEN 3 THEN CASE WHEN n=1 THEN 'Начать со штатной проверки' ELSE 'Сформулировать безопасный Ответ пользователя своими словами' END
         ELSE 'Вести диалог, сохраняя безопасный порядок сделки' END,
       CASE n WHEN 1 THEN s.opening WHEN 2 THEN s.deceptive_request ELSE s.pressure END,
       CASE WHEN l.level_number>=3 THEN 'Risk: '||s.risk_type||'. Оцени текущий Ответ пользователя: '||s.safe_action||' Опасное действие: '||s.unsafe_action END,
       CASE n WHEN 1 THEN s.deceptive_request ELSE s.pressure END,100
FROM chats c
JOIN migration_000010_new_chats m ON m.id=c.id
JOIN levels l ON l.id=c.level_id
JOIN migration_000010_content_seed s ON s.topic_slug=c.product_context->>'topic'
CROSS JOIN LATERAL generate_series(1,CASE l.level_number WHEN 1 THEN 3 WHEN 2 THEN 2 WHEN 3 THEN 3 ELSE 6 END)n;

INSERT INTO chat_options(step_id,option_text,counterparty_reaction,explanation,points,sort_order)
SELECT st.id,
       CASE l.level_number
         WHEN 1 THEN CASE st.step_number
           WHEN 1 THEN (ARRAY[s.safe_action,s.cautious_action,s.unsafe_action])[o.n]
           WHEN 2 THEN (ARRAY[s.verification,s.cautious_action,s.unsafe_action])[o.n]
           ELSE (ARRAY[s.summary,s.cautious_action,s.unsafe_action])[o.n] END
         WHEN 2 THEN CASE st.step_number
           WHEN 1 THEN (ARRAY['Наиболее безопасно: '||s.safe_action,s.cautious_action||' и затем решить.',s.unsafe_action||' после короткой проверки.'])[o.n]
           ELSE (ARRAY['Проверю самостоятельно: '||s.verification,'Попрошу собеседника подтвердить условия и затем проверю сам.',s.unsafe_action||', если сообщение выглядит убедительно.'])[o.n] END
         ELSE (ARRAY['Начну со штатного действия: '||s.safe_action,'Сначала '||lower(s.cautious_action),s.unsafe_action])[o.n] END,
       CASE WHEN (l.level_number=1 AND st.step_number<3) OR (l.level_number=2 AND st.step_number<2) OR l.level_number=3
         THEN CASE st.step_number
           WHEN 1 THEN (ARRAY['Проверяйте, но я всё равно предлагаю оформить быстрее.','Хорошо, но для продолжения всё равно понадобится подтверждение.','Тогда переходите к оформлению по моей инструкции.'])[o.n]
           ELSE (ARRAY['Собеседник не принимает отказ и начинает торопить.','Собеседник усиливает срочность.','Собеседник требует завершить действие немедленно.'])[o.n] END END,
       CASE o.n WHEN 1 THEN s.safe_action WHEN 2 THEN 'Действие снижает часть риска, но не завершает самостоятельную проверку.' ELSE 'Опасное действие: '||s.unsafe_action END,
       (ARRAY[100,50,0])[o.n],o.n
FROM chat_steps st
JOIN chats c ON c.id=st.chat_id
JOIN migration_000010_new_chats m ON m.id=c.id
JOIN levels l ON l.id=c.level_id
JOIN migration_000010_content_seed s ON s.topic_slug=c.product_context->>'topic'
CROSS JOIN generate_series(1,3)o(n)
WHERE l.level_number IN (1,2) OR (l.level_number=3 AND st.step_number=1);

DROP TABLE migration_000010_content_seed;
