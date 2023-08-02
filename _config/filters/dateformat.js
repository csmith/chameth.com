import {DateTime} from 'luxon';

export default (date) =>
    DateTime
        .fromJSDate(date)
        .toLocaleString(DateTime.DATE_MED);