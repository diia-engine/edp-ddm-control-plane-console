import i18n from "../localization";

export const getStatusTitle = (status: string): string => {
    switch (`status-${status}`) {
        case "status-active":
        case "status-SUCCESS":
        case "status-ok":
            return i18n.global.t('domains.changes.statuses.active');
        case "status-failed":
        case "status-failure":
        case "status-FAILURE":
        case "status-ABORTED":
            return i18n.global.t('domains.changes.statuses.error');
        case "status-inactive":
            return i18n.global.t('domains.changes.statuses.inProgress');
        case "status-disabled":
            return i18n.global.t('domains.changes.statuses.disabled');
        default: 
            return '';
    }
};

export const getImageUrl = (name: string): string => {
    return new URL(`../assets/img/${name.toLocaleLowerCase()}.png`, import.meta.url).href;
};

export const getGerritURL = (url: string): string => {
    //TODO: need to be link to specific repo
    return `${url}/dashboard/self`;
};

export const getJenkinsURL = (url: string, codebaseName: string, branchName: string): string => {
    return `${url}/job/${codebaseName}/view/${branchName.toLocaleUpperCase()}`;
};