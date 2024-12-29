.PHONY: all cpu memoria filesystem kernel

all: cpu filesystem kernel memoria

cpu:
	@cd cpu && mkdir -p bin && go build -o bin/cpu && ./bin/cpu $(ENV) $(C)

memoria:
	@cd memoria && mkdir -p bin && go build -o bin/memoria && ./bin/memoria $(ENV) $(C)

filesystem:
	@cd filesystem && mkdir -p bin && go build -o bin/filesystem && ./bin/filesystem $(ENV) $(C)

kernel:
	@cd kernel && mkdir -p bin && go build -o bin/kernel && ./bin/kernel $(ENV) $(C) $(COD) $(TAM)


clean:
	@rm -rf cpu/bin memoria/bin filesystem/bin kernel/bin