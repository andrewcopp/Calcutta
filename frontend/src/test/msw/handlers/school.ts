import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

export const schoolHandlers = [
  http.get(`${BASE}/schools`, () => {
    return HttpResponse.json({
      items: [
        { id: 'sch-1', name: 'Duke' },
        { id: 'sch-2', name: 'UNC' },
      ],
    });
  }),
];
