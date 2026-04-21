import request from './request';
import { RequestOptions, useApiFetch } from '@/hooks/useRequest';

/**
 * 可用于各模块定义各自api请求
 * T 请求返回的数据类型
 * P 请求参数类型
 */
class Api<T = any, P = any> {
    /**
     * 请求url
     */
    url: string;

    /**
     * 请求方法
     */
    method: string;

    /**
     * 请求前处理函数
     * param1: param请求参数
     */
    beforeHandler: Function;

    constructor(url: string, method: string) {
        this.url = url;
        this.method = method;
    }

    /**
     * 设置请求前处理回调函数
     * @param func 请求前处理器
     * @returns this
     */
    withBeforeHandler(func: Function) {
        this.beforeHandler = func;
        return this;
    }

    /**
     * 获取权限的完整url
     */
    getUrl() {
        return request.getApiUrl(this.url);
    }

    /**
     * 响应式使用该api
     * @param param 请求参数
     * @param reqOptions 其他可选值
     * @returns
     */
    useApi(param?: P, reqOptions?: RequestOptions) {
        return useApiFetch<T, P>(this, param, reqOptions);
    }

    /**
     * fetch 请求对应的该api
     * @param {Object} param 请求该api的参数
     */
    async request(param?: P, options: any = {}): Promise<T> {
        const { execute, data } = this.useApi(param, options);
        const res = await execute();
        return (data.value as T) || (res as T);
    }

    /**
     * xhr 请求对应的该api
     * @param {Object} param 请求该api的参数
     */
    async xhrReq(param: any = null, options: any = {}): Promise<T> {
        if (this.beforeHandler) {
            await this.beforeHandler(param);
        }
        return request.xhrReq(this.method, this.url, param, options);
    }

    /**    静态方法     **/

    /**
     * 静态工厂，返回Api对象，并设置url与method属性
     * @param url url
     * @param method 请求方法(get,post,put,delete...)
     */
    static create<T = any, P = any>(url: string, method: string): Api<T> {
        return new Api<T, P>(url, method);
    }

    /**
     * 创建get api
     * @param url url
     */
    static newGet<T = any, P = any>(url: string): Api<T, P> {
        return Api.create<T, P>(url, 'get');
    }

    /**
     * new post api
     * @param url url
     */
    static newPost<T = any, P = any>(url: string): Api<T, P> {
        return Api.create<T, P>(url, 'post');
    }

    /**
     * new put api
     * @param url url
     */
    static newPut<T = any, P = any>(url: string): Api<T, P> {
        return Api.create<T, P>(url, 'put');
    }

    /**
     * new delete api
     * @param url url
     */
    static newDelete<T = any, P = any>(url: string): Api<T, P> {
        return Api.create<T, P>(url, 'delete');
    }
}

export default Api;

export class PageRes {
    list: any[] = [];
    total: number = 0;
}
