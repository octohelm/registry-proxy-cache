WORKING_DIR ?= .
COMPONENT ?= default
NAMESPACE ?= default
NAME ?= $(COMPONENT)
ENV ?= default
DEBUG ?= 1

KUBECONFIG = ~/.kube/config--$(ENV).yaml

HELM ?= KUBECONFIG=$(KUBECONFIG) helm upgrade --install --create-namespace
ifeq ($(DEBUG),1)
	HELM = helm template --dependency-update
endif

VALUES_FILE = $(WORKING_DIR)/values.yaml

ifneq ($(ENV),default)
	VALUES_FILE := $(WORKING_DIR)/values--$(ENV).yaml
endif

COMPONENT_VALUES_FILE = $(WORKING_DIR)/components/$(COMPONENT)/values--$(ENV).yaml

info:
	KUBECONFIG=$(KUBECONFIG) kubectl version

gen:
	if [[ -f $(COMPONENT_VALUES_FILE) ]]; then rm $(COMPONENT_VALUES_FILE); fi;
	if [[ -f $(WORKING_DIR)/templates/$(COMPONENT).yaml ]]; then helm template --values=$(VALUES_FILE) --show-only templates/$(COMPONENT).yaml $(WORKING_DIR) > $(COMPONENT_VALUES_FILE); fi;

apply: gen
	if [[ -f $(COMPONENT_VALUES_FILE) ]]; then \
  		$(HELM) --namespace=$(NAMESPACE) --values=$(COMPONENT_VALUES_FILE) $(NAME) $(WORKING_DIR)/components/$(COMPONENT); \
  	else \
		$(HELM) --namespace=$(NAMESPACE) $(NAME) $(WORKING_DIR)/components/$(COMPONENT); \
  	fi;
