
# Базовый образ RabbitMQ с веб-интерфейсом
FROM rabbitmq:4.2-management

# Определяем версию плагина
ENV DELAYED_EXCHANGE_VERSION=4.2.0

# Настройка учётных данных
ENV RABBITMQ_DEFAULT_USER=guest
ENV RABBITMQ_DEFAULT_PASS=guest

# Скачиваем плагин
ADD https://github.com/rabbitmq/rabbitmq-delayed-message-exchange/releases/download/v${DELAYED_EXCHANGE_VERSION}/rabbitmq_delayed_message_exchange-${DELAYED_EXCHANGE_VERSION}.ez /opt/rabbitmq/plugins/

# Выставляем права
RUN chown rabbitmq:rabbitmq /opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-${DELAYED_EXCHANGE_VERSION}.ez \
    && chmod 644 /opt/rabbitmq/plugins/rabbitmq_delayed_message_exchange-${DELAYED_EXCHANGE_VERSION}.ez

# Включаем плагин
RUN rabbitmq-plugins enable --offline rabbitmq_delayed_message_exchange

# Пробрасываем порты
EXPOSE 5672 15672

# Запуск сервера
CMD ["rabbitmq-server"]