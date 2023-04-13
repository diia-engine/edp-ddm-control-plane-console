import moment from "moment";

export const getFormattedDate = (date: string): string => {
    return moment(date).utc().format('DD.MM.YYYY H:mm');
};