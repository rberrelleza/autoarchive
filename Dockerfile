FROM golang:onbuild
ENV SCHEDULER_MODE local
ENV REDIS_URL redis://127.0.0.1:6379
ENV BASE_URL http://localhost:${PORT}

