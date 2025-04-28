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

export const semVerComparator = (a: string, b: string): number => {
  // Split the version strings into arrays of numbers
  const versionA = a.split('.').map(Number);
  const versionB = b.split('.').map(Number);

  // Compare each segment of the version
  for (let i = 0; i < Math.max(versionA.length, versionB.length); i++) {
    const partA = versionA[i] || 0; // Default to 0 if the part does not exist
    const partB = versionB[i] || 0; // Default to 0 if the part does not exist

    if (partA !== partB) {
      return partA - partB; // Compare numerically
    }
  }

  // If all parts are equal, the versions are identical
  return 0;
};