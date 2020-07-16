FROM golang
WORKDIR /karst
COPY . .
RUN ./install.sh
CMD karst init $INIT_ARGS && karst daemon
