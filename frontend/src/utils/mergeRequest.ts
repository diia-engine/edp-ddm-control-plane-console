import i18n from "../localization";

export const getMergeRequestName = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];
    if (target === "external-reg") {
        return metadata.annotations["ext-reg/name"];
    }

    if (target == "registry-version-update") {
        return i18n.global.t('domains.changes.mergeRequest.names.updatingRegistryVersion');
    }

    if (target == "publicAPI-reg") {
        return metadata.labels['publicAPI-reg-name'];
    }

    if (target == "edit-registry" || target == "trembita-registry-update") {
        return i18n.global.t('domains.changes.mergeRequest.names.editingRegistry');
    }

    return metadata.name;
};

export const isRegistryUpdateMrOpen = (mergeRequest: any): boolean => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];
    return mergeRequest.status.value === "NEW" && target == "registry-version-update";
};

export const getMergeRequestAction = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];
    if (target == "external-reg") {
        const res = metadata.labels["console/sub-target"];
        if (res) {
            switch (`mre-action-${res}`) {
                case "mre-action-disable":
                    return i18n.global.t('domains.changes.mergeRequest.actions.disable');
                case "mre-action-enable":
                    return i18n.global.t('domains.changes.mergeRequest.actions.enable');
                case "mre-action-creation":
                    return i18n.global.t('domains.changes.mergeRequest.actions.creation');
                case "mre-action-deletion":
                    return i18n.global.t('domains.changes.mergeRequest.actions.deletion');
            }
        }
    }

    if (target == "publicAPI-reg") {
        const res = metadata.labels["console/sub-target"];
        if (res) {
            switch (res) {
                case "edition":
                    return i18n.global.t('domains.changes.mergeRequest.actions.edition');
                case "disable":
                    return i18n.global.t('domains.changes.mergeRequest.actions.disable');
                case "enable":
                    return i18n.global.t('domains.changes.mergeRequest.actions.enable');
                case "creation":
                    return i18n.global.t('domains.changes.mergeRequest.actions.creation');
                case "deletion":
                    return i18n.global.t('domains.changes.mergeRequest.actions.removing');
            }
        }
    }

    if (target == "registry-version-update") {
        let sourceBranch = mergeRequest.spec.sourceBranch;
        if (sourceBranch === "") {
            sourceBranch = metadata.labels["console/source-branch"];
        }

        return i18n.global.t('domains.changes.mergeRequest.actions.versionUpdate', { sourceBranch });
    }

    return "-";
};

export const getMergeRequestPlatformAction = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];

	if (target === "cluster-admins") {
		return i18n.global.t('domains.changes.mergeRequest.platformActions.updatingAdministrators');
	}

    if (target === "cluster-cidr") {
		return i18n.global.t('domains.changes.mergeRequest.platformActions.accessRestrictions');
	}
  

    if (target === "cluster-keycloak-dns") {
		return i18n.global.t('domains.changes.mergeRequest.platformActions.editingDNSKeycloak');
	}

	let sourceBranch = mergeRequest.spec.sourceBranch;
	if (sourceBranch === "") {
		sourceBranch = ["console/source-branch"];
	}

	if (target == "cluster-update") {
		return i18n.global.t('domains.changes.mergeRequest.platformActions.updatePlatform', { sourceBranch });
	}

	return "-";
};

export const mrIsInProgress = (mergeRequest: any) => {
    const { metadata } = mergeRequest;
    return (metadata?.labels?.["console/action"] == "branch-merge" && mergeRequest.spec.sourceBranch == "") || mergeRequest.status.value == "" ||
        (mergeRequest.status.value == "sourceBranch or changesConfigMap must be specified" && mergeRequest.spec.sourceBranch != "");
};

export const getMergeRequestStatus = (mergeRequest: any) => {
    if (mrIsInProgress(mergeRequest)) {
        return i18n.global.t('domains.changes.mergeRequest.statuses.inProgress');
    }

    if (mergeRequest.status.value === "") {
        return "-";
    }

    switch (mergeRequest.status.value) {
        case "NEW":
            return i18n.global.t('domains.changes.mergeRequest.statuses.new');
        case "ABANDONED":
            return i18n.global.t('domains.changes.mergeRequest.statuses.abandoned');
        case "MERGED":
            return i18n.global.t('domains.changes.mergeRequest.statuses.merged');
    }
};