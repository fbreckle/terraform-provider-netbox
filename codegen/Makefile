all: resources provider scaffold

spec_gen.json:
	tfplugingen-openapi generate --config generator_config.yml --output spec_gen.json netbox-3.6.yaml

.PHONY: code
resources: spec_gen.json
	tfplugingen-framework generate resources --input spec_gen.json --output ../internal/generate

.PHONY: provider
provider:
	tfplugingen-framework generate provider --input provider.json --output ../internal/generate

.PHONY: scaffold
scaffold:
	mkdir -p ../internal/provider
	tfplugingen-framework scaffold resource --name netbox --output-dir ../internal/provider

.PHONY: clean
clean:
	rm -rf ../internal/provider
