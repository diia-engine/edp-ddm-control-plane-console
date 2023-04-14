export const getStatus = (status: string): string => {
    switch (`status-${status}`) {
        case "status-active":
        case "status-SUCCESS":
        case "status-ok":
            return "Активний";
        case "status-failed":
        case "status-failure":
        case "status-FAILURE":
            return "Помилка";
        case "status-inactive":
            return "В обробці";
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