

export const getMergeRequestName = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];
    if (target === "external-reg") {
        return metadata.annotations["ext-reg/name"];
    }

    if (target == "registry-version-update") {
        return "Оновлення версії реєстру";
    }

    if (target == "edit-registry" || target == "trembita-registry-update") {
        return "Редагування реєстру";
    }

    return metadata.name;
};

export const getMergeRequestAction = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];
    if (target == "external-reg") {
        const res = metadata.labels["console/sub-target"];
        if (res) {
            switch (`mre-action-${res}`) {
                case "mre-action-disable":
                    return "Заблокування";
                case "mre-action-enable":
                    return "Розблокування";
                case "mre-action-creation":
                    return "Створення";
                case "mre-action-deletion":
                    return "Скасування";
            }
        }
    }

    if (target == "registry-version-update") {
        let sourceBranch = mergeRequest.spec.sourceBranch;
        if (sourceBranch === "") {
            sourceBranch = metadata.labels["console/source-branch"];
        }

        return `Оновлення реєстру до ${sourceBranch}`;
    }

    return "-";
};

export const getMergeRequestPlatformAction = (mergeRequest: any): string => {
    const { metadata } = mergeRequest;
    const target = metadata?.labels?.["console/target"];

	if (target === "cluster-admins") {
		return "Оновлення адміністраторів платформи";
	}

    if (target === "cluster-cidr") {
		return "Обмеження доступу";
	}
  

    if (target === "cluster-keycloak-dns") {
		return "Редагування DNS Keycloak";
	}

	let sourceBranch = mergeRequest.spec.sourceBranch;
	if (sourceBranch === "") {
		sourceBranch = ["console/source-branch"];
	}

	if (target == "cluster-update") {
		return `Оновлення платформи до ${sourceBranch}`;
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
        return "У процесі виконання";
    }

    if (mergeRequest.status.value === "") {
        return "-";
    }

    switch (mergeRequest.status.value) {
        case "NEW":
            return "Новий";
        case "ABANDONED":
            return "Відхилено";
        case "MERGED":
            return "Підтверджено";
    }
};