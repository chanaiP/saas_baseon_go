--
-- PostgreSQL database dump
--

-- Dumped from database version 16.13
-- Dumped by pg_dump version 16.13

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: app_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_user (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    employee_no character varying(64) NOT NULL,
    account character varying(64) NOT NULL,
    password_hash character varying(255) NOT NULL,
    name character varying(128) NOT NULL,
    phone character varying(32),
    email character varying(128),
    avatar_url text,
    status bigint DEFAULT 1 NOT NULL,
    is_platform_admin boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    company_id bigint,
    department_id bigint
);


--
-- Name: app_user_department; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_user_department (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    department_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: app_user_department_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.app_user_department_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: app_user_department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.app_user_department_id_seq OWNED BY public.app_user_department.id;


--
-- Name: app_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.app_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: app_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.app_user_id_seq OWNED BY public.app_user.id;


--
-- Name: app_user_position; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_user_position (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    position_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: app_user_position_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.app_user_position_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: app_user_position_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.app_user_position_id_seq OWNED BY public.app_user_position.id;


--
-- Name: audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_log (
    id bigint NOT NULL,
    tenant_id bigint,
    user_id bigint,
    module character varying(64) NOT NULL,
    action character varying(32) NOT NULL,
    summary character varying(500) NOT NULL,
    detail text,
    ip character varying(64),
    created_at timestamp with time zone NOT NULL
);


--
-- Name: audit_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_log_id_seq OWNED BY public.audit_log.id;


--
-- Name: business_unit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_unit (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    name character varying(200) NOT NULL,
    code character varying(64) NOT NULL,
    bu_type character varying(32),
    status bigint DEFAULT 1 NOT NULL,
    billing_enabled boolean DEFAULT false NOT NULL,
    statistic_enabled boolean DEFAULT false NOT NULL,
    remark text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: business_unit_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.business_unit_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: business_unit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.business_unit_id_seq OWNED BY public.business_unit.id;


--
-- Name: business_unit_org_map; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_unit_org_map (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    business_unit_id bigint NOT NULL,
    org_id bigint NOT NULL,
    org_type character varying(32) NOT NULL,
    scope_type character varying(32) NOT NULL,
    priority bigint DEFAULT 0 NOT NULL,
    effective_start timestamp with time zone,
    effective_end timestamp with time zone,
    status bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: business_unit_org_map_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.business_unit_org_map_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: business_unit_org_map_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.business_unit_org_map_id_seq OWNED BY public.business_unit_org_map.id;


--
-- Name: business_unit_scope; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_unit_scope (
    id bigint NOT NULL,
    role_permission_id bigint NOT NULL,
    business_unit_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: business_unit_scope_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.business_unit_scope_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: business_unit_scope_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.business_unit_scope_id_seq OWNED BY public.business_unit_scope.id;


--
-- Name: dict_item; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dict_item (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    dict_type_id bigint NOT NULL,
    label character varying(200) NOT NULL,
    value character varying(200) NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: dict_item_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.dict_item_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dict_item_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.dict_item_id_seq OWNED BY public.dict_item.id;


--
-- Name: dict_type; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dict_type (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(200) NOT NULL,
    remark character varying(500),
    scope character varying(32) NOT NULL,
    tenant_editable boolean DEFAULT false NOT NULL,
    is_platform_only boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: dict_type_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.dict_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dict_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.dict_type_id_seq OWNED BY public.dict_type.id;


--
-- Name: login_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.login_log (
    id bigint NOT NULL,
    tenant_id bigint,
    user_id bigint,
    account character varying(128) NOT NULL,
    success boolean NOT NULL,
    message character varying(500),
    ip character varying(64),
    user_agent character varying(500),
    created_at timestamp with time zone NOT NULL
);


--
-- Name: login_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.login_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: login_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.login_log_id_seq OWNED BY public.login_log.id;


--
-- Name: org_node; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.org_node (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    node_type character varying(32) NOT NULL,
    parent_id bigint,
    company_id bigint,
    name character varying(200) NOT NULL,
    code character varying(64),
    company_type character varying(32),
    status bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: org_node_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.org_node_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: org_node_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.org_node_id_seq OWNED BY public.org_node.id;


--
-- Name: permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    parent_id bigint,
    name character varying(128) NOT NULL,
    path character varying(255) NOT NULL,
    perm_type bigint NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    visible boolean DEFAULT true NOT NULL,
    is_platform_only boolean DEFAULT false NOT NULL,
    is_package_feature boolean DEFAULT true NOT NULL,
    tenant_editable boolean DEFAULT false NOT NULL,
    tenant_edit_scope character varying(64),
    feature_code character varying(128),
    feature_type character varying(32),
    data_perm_mode character varying(16) DEFAULT 'ORG'::character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    data_scope character varying(32)
);


--
-- Name: permission_custom_department; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission_custom_department (
    id bigint NOT NULL,
    permission_id bigint NOT NULL,
    department_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: permission_custom_department_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permission_custom_department_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permission_custom_department_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permission_custom_department_id_seq OWNED BY public.permission_custom_department.id;


--
-- Name: permission_custom_user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission_custom_user (
    id bigint NOT NULL,
    permission_id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: permission_custom_user_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permission_custom_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permission_custom_user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permission_custom_user_id_seq OWNED BY public.permission_custom_user.id;


--
-- Name: permission_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permission_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permission_id_seq OWNED BY public.permission.id;


--
-- Name: position; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."position" (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    position_type_id bigint NOT NULL,
    name character varying(200) NOT NULL,
    code character varying(64) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: position_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.position_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: position_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.position_id_seq OWNED BY public."position".id;


--
-- Name: position_type; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.position_type (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    name character varying(200) NOT NULL,
    code character varying(64) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: position_type_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.position_type_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: position_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.position_type_id_seq OWNED BY public.position_type.id;


--
-- Name: role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(200) NOT NULL,
    status bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    description character varying(500)
);


--
-- Name: role_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.role_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.role_id_seq OWNED BY public.role.id;


--
-- Name: role_permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permission (
    id bigint NOT NULL,
    role_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    data_scope_override character varying(32),
    custom_company_ids_json text,
    custom_department_ids_json text,
    custom_user_ids_json text,
    custom_business_unit_ids_json text,
    bu_data_access_mode character varying(32)
);


--
-- Name: role_permission_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.role_permission_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: role_permission_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.role_permission_id_seq OWNED BY public.role_permission.id;


--
-- Name: saas_feature; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_feature (
    id bigint NOT NULL,
    feature_code character varying(100) NOT NULL,
    feature_name character varying(100) NOT NULL,
    feature_type character varying(32) NOT NULL,
    parent_id bigint DEFAULT 0 NOT NULL,
    menu_id bigint,
    api_method character varying(20),
    api_path character varying(255),
    service_key character varying(100),
    status bigint DEFAULT 1 NOT NULL,
    description character varying(500),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: saas_feature_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_feature_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_feature_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_feature_id_seq OWNED BY public.saas_feature.id;


--
-- Name: saas_plan; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_plan (
    id bigint NOT NULL,
    plan_code character varying(64) NOT NULL,
    plan_name character varying(100) NOT NULL,
    plan_type character varying(32) NOT NULL,
    billing_cycle character varying(32) NOT NULL,
    price numeric(12,2) DEFAULT 0 NOT NULL,
    status bigint DEFAULT 1 NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    sort_order bigint DEFAULT 0 NOT NULL,
    description character varying(500),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: saas_plan_feature; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_plan_feature (
    id bigint NOT NULL,
    plan_id bigint NOT NULL,
    feature_id bigint NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: saas_plan_feature_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_plan_feature_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_plan_feature_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_plan_feature_id_seq OWNED BY public.saas_plan_feature.id;


--
-- Name: saas_plan_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_plan_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_plan_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_plan_id_seq OWNED BY public.saas_plan.id;


--
-- Name: saas_plan_quota; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_plan_quota (
    id bigint NOT NULL,
    plan_id bigint NOT NULL,
    quota_id bigint NOT NULL,
    quota_value bigint DEFAULT 0 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: saas_plan_quota_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_plan_quota_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_plan_quota_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_plan_quota_id_seq OWNED BY public.saas_plan_quota.id;


--
-- Name: saas_quota; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_quota (
    id bigint NOT NULL,
    quota_code character varying(100) NOT NULL,
    quota_name character varying(100) NOT NULL,
    quota_type character varying(32) NOT NULL,
    period_type character varying(32),
    unit character varying(32),
    status bigint DEFAULT 1 NOT NULL,
    description character varying(500),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: saas_quota_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_quota_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_quota_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_quota_id_seq OWNED BY public.saas_quota.id;


--
-- Name: sys_param; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_param (
    id bigint NOT NULL,
    tenant_id bigint DEFAULT 1 NOT NULL,
    param_key character varying(128) NOT NULL,
    param_value text NOT NULL,
    remark character varying(500) DEFAULT ''::character varying NOT NULL,
    value_type character varying(32) DEFAULT 'string'::character varying NOT NULL,
    tenant_editable boolean DEFAULT true NOT NULL,
    is_platform_only boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: sys_param_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_param_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_param_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_param_id_seq OWNED BY public.sys_param.id;


--
-- Name: system_param; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.system_param (
    id bigint NOT NULL,
    param_key character varying(128) NOT NULL,
    param_value text NOT NULL,
    remark text DEFAULT ''::text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: system_param_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.system_param_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: system_param_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.system_param_id_seq OWNED BY public.system_param.id;


--
-- Name: tenant; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant (
    id bigint NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(200) NOT NULL,
    status bigint DEFAULT 1 NOT NULL,
    is_platform boolean DEFAULT false NOT NULL,
    brand_name character varying(128),
    footer_text text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    start_date date,
    expire_date date,
    max_companies bigint DEFAULT 0 NOT NULL,
    max_users bigint DEFAULT 0 NOT NULL,
    contact_name character varying(100),
    contact_phone character varying(32),
    is_platform_tenant boolean DEFAULT false NOT NULL,
    brand_display_name character varying(128),
    brand_logo_data text,
    brand_footer_text character varying(256)
);


--
-- Name: tenant_dict_item_override; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_dict_item_override (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    dict_item_id bigint NOT NULL,
    custom_label character varying(200),
    custom_value character varying(200),
    enabled boolean,
    sort_order bigint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_dict_item_override_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_dict_item_override_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_dict_item_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_dict_item_override_id_seq OWNED BY public.tenant_dict_item_override.id;


--
-- Name: tenant_feature_override; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_feature_override (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    feature_id bigint NOT NULL,
    enabled boolean NOT NULL,
    reason character varying(500),
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_feature_override_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_feature_override_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_feature_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_feature_override_id_seq OWNED BY public.tenant_feature_override.id;


--
-- Name: tenant_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_id_seq OWNED BY public.tenant.id;


--
-- Name: tenant_menu_override; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_menu_override (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    permission_id bigint NOT NULL,
    custom_name character varying(200),
    custom_icon character varying(100),
    enabled boolean,
    visible boolean,
    sort_order bigint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_menu_override_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_menu_override_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_menu_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_menu_override_id_seq OWNED BY public.tenant_menu_override.id;


--
-- Name: tenant_param_value; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_param_value (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    param_id bigint NOT NULL,
    param_value text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_param_value_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_param_value_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_param_value_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_param_value_id_seq OWNED BY public.tenant_param_value.id;


--
-- Name: tenant_quota_override; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_quota_override (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    quota_id bigint NOT NULL,
    quota_value bigint NOT NULL,
    reason character varying(500),
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_quota_override_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_quota_override_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_quota_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_quota_override_id_seq OWNED BY public.tenant_quota_override.id;


--
-- Name: tenant_quota_usage; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_quota_usage (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    quota_code character varying(100) NOT NULL,
    used_value bigint DEFAULT 0 NOT NULL,
    limit_value bigint DEFAULT 0 NOT NULL,
    period_type character varying(32),
    period_key character varying(32) NOT NULL,
    last_refresh_time timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_quota_usage_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_quota_usage_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_quota_usage_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_quota_usage_id_seq OWNED BY public.tenant_quota_usage.id;


--
-- Name: tenant_subscription; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_subscription (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    plan_id bigint NOT NULL,
    subscription_status character varying(32) NOT NULL,
    start_time timestamp with time zone NOT NULL,
    end_time timestamp with time zone,
    trial_end_time timestamp with time zone,
    auto_renew boolean DEFAULT false NOT NULL,
    frozen_reason character varying(500),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: tenant_subscription_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_subscription_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_subscription_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_subscription_id_seq OWNED BY public.tenant_subscription.id;


--
-- Name: user_preference; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_preference (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    pref_key character varying(64) NOT NULL,
    pref_value text,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: user_preference_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_preference_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_preference_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_preference_id_seq OWNED BY public.user_preference.id;


--
-- Name: user_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_role (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: user_role_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_role_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_role_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_role_id_seq OWNED BY public.user_role.id;


--
-- Name: app_user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user ALTER COLUMN id SET DEFAULT nextval('public.app_user_id_seq'::regclass);


--
-- Name: app_user_department id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user_department ALTER COLUMN id SET DEFAULT nextval('public.app_user_department_id_seq'::regclass);


--
-- Name: app_user_position id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user_position ALTER COLUMN id SET DEFAULT nextval('public.app_user_position_id_seq'::regclass);


--
-- Name: audit_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log ALTER COLUMN id SET DEFAULT nextval('public.audit_log_id_seq'::regclass);


--
-- Name: business_unit id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit ALTER COLUMN id SET DEFAULT nextval('public.business_unit_id_seq'::regclass);


--
-- Name: business_unit_org_map id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit_org_map ALTER COLUMN id SET DEFAULT nextval('public.business_unit_org_map_id_seq'::regclass);


--
-- Name: business_unit_scope id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit_scope ALTER COLUMN id SET DEFAULT nextval('public.business_unit_scope_id_seq'::regclass);


--
-- Name: dict_item id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dict_item ALTER COLUMN id SET DEFAULT nextval('public.dict_item_id_seq'::regclass);


--
-- Name: dict_type id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dict_type ALTER COLUMN id SET DEFAULT nextval('public.dict_type_id_seq'::regclass);


--
-- Name: login_log id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_log ALTER COLUMN id SET DEFAULT nextval('public.login_log_id_seq'::regclass);


--
-- Name: org_node id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.org_node ALTER COLUMN id SET DEFAULT nextval('public.org_node_id_seq'::regclass);


--
-- Name: permission id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission ALTER COLUMN id SET DEFAULT nextval('public.permission_id_seq'::regclass);


--
-- Name: permission_custom_department id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_custom_department ALTER COLUMN id SET DEFAULT nextval('public.permission_custom_department_id_seq'::regclass);


--
-- Name: permission_custom_user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_custom_user ALTER COLUMN id SET DEFAULT nextval('public.permission_custom_user_id_seq'::regclass);


--
-- Name: position id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."position" ALTER COLUMN id SET DEFAULT nextval('public.position_id_seq'::regclass);


--
-- Name: position_type id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.position_type ALTER COLUMN id SET DEFAULT nextval('public.position_type_id_seq'::regclass);


--
-- Name: role id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role ALTER COLUMN id SET DEFAULT nextval('public.role_id_seq'::regclass);


--
-- Name: role_permission id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permission ALTER COLUMN id SET DEFAULT nextval('public.role_permission_id_seq'::regclass);


--
-- Name: saas_feature id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_feature ALTER COLUMN id SET DEFAULT nextval('public.saas_feature_id_seq'::regclass);


--
-- Name: saas_plan id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan ALTER COLUMN id SET DEFAULT nextval('public.saas_plan_id_seq'::regclass);


--
-- Name: saas_plan_feature id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan_feature ALTER COLUMN id SET DEFAULT nextval('public.saas_plan_feature_id_seq'::regclass);


--
-- Name: saas_plan_quota id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan_quota ALTER COLUMN id SET DEFAULT nextval('public.saas_plan_quota_id_seq'::regclass);


--
-- Name: saas_quota id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_quota ALTER COLUMN id SET DEFAULT nextval('public.saas_quota_id_seq'::regclass);


--
-- Name: sys_param id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_param ALTER COLUMN id SET DEFAULT nextval('public.sys_param_id_seq'::regclass);


--
-- Name: system_param id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_param ALTER COLUMN id SET DEFAULT nextval('public.system_param_id_seq'::regclass);


--
-- Name: tenant id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant ALTER COLUMN id SET DEFAULT nextval('public.tenant_id_seq'::regclass);


--
-- Name: tenant_dict_item_override id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_dict_item_override ALTER COLUMN id SET DEFAULT nextval('public.tenant_dict_item_override_id_seq'::regclass);


--
-- Name: tenant_feature_override id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_feature_override ALTER COLUMN id SET DEFAULT nextval('public.tenant_feature_override_id_seq'::regclass);


--
-- Name: tenant_menu_override id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_menu_override ALTER COLUMN id SET DEFAULT nextval('public.tenant_menu_override_id_seq'::regclass);


--
-- Name: tenant_param_value id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_param_value ALTER COLUMN id SET DEFAULT nextval('public.tenant_param_value_id_seq'::regclass);


--
-- Name: tenant_quota_override id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_quota_override ALTER COLUMN id SET DEFAULT nextval('public.tenant_quota_override_id_seq'::regclass);


--
-- Name: tenant_quota_usage id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_quota_usage ALTER COLUMN id SET DEFAULT nextval('public.tenant_quota_usage_id_seq'::regclass);


--
-- Name: tenant_subscription id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_subscription ALTER COLUMN id SET DEFAULT nextval('public.tenant_subscription_id_seq'::regclass);


--
-- Name: user_preference id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_preference ALTER COLUMN id SET DEFAULT nextval('public.user_preference_id_seq'::regclass);


--
-- Name: user_role id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_role ALTER COLUMN id SET DEFAULT nextval('public.user_role_id_seq'::regclass);


--
-- Name: app_user_department app_user_department_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user_department
    ADD CONSTRAINT app_user_department_pkey PRIMARY KEY (id);


--
-- Name: app_user app_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user
    ADD CONSTRAINT app_user_pkey PRIMARY KEY (id);


--
-- Name: app_user_position app_user_position_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_user_position
    ADD CONSTRAINT app_user_position_pkey PRIMARY KEY (id);


--
-- Name: audit_log audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_log
    ADD CONSTRAINT audit_log_pkey PRIMARY KEY (id);


--
-- Name: business_unit_org_map business_unit_org_map_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit_org_map
    ADD CONSTRAINT business_unit_org_map_pkey PRIMARY KEY (id);


--
-- Name: business_unit business_unit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit
    ADD CONSTRAINT business_unit_pkey PRIMARY KEY (id);


--
-- Name: business_unit_scope business_unit_scope_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_unit_scope
    ADD CONSTRAINT business_unit_scope_pkey PRIMARY KEY (id);


--
-- Name: dict_item dict_item_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dict_item
    ADD CONSTRAINT dict_item_pkey PRIMARY KEY (id);


--
-- Name: dict_type dict_type_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dict_type
    ADD CONSTRAINT dict_type_pkey PRIMARY KEY (id);


--
-- Name: login_log login_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_log
    ADD CONSTRAINT login_log_pkey PRIMARY KEY (id);


--
-- Name: org_node org_node_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.org_node
    ADD CONSTRAINT org_node_pkey PRIMARY KEY (id);


--
-- Name: permission_custom_department permission_custom_department_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_custom_department
    ADD CONSTRAINT permission_custom_department_pkey PRIMARY KEY (id);


--
-- Name: permission_custom_user permission_custom_user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_custom_user
    ADD CONSTRAINT permission_custom_user_pkey PRIMARY KEY (id);


--
-- Name: permission permission_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT permission_pkey PRIMARY KEY (id);


--
-- Name: position position_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."position"
    ADD CONSTRAINT position_pkey PRIMARY KEY (id);


--
-- Name: position_type position_type_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.position_type
    ADD CONSTRAINT position_type_pkey PRIMARY KEY (id);


--
-- Name: role_permission role_permission_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permission
    ADD CONSTRAINT role_permission_pkey PRIMARY KEY (id);


--
-- Name: role role_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role_pkey PRIMARY KEY (id);


--
-- Name: saas_feature saas_feature_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_feature
    ADD CONSTRAINT saas_feature_pkey PRIMARY KEY (id);


--
-- Name: saas_plan_feature saas_plan_feature_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan_feature
    ADD CONSTRAINT saas_plan_feature_pkey PRIMARY KEY (id);


--
-- Name: saas_plan saas_plan_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan
    ADD CONSTRAINT saas_plan_pkey PRIMARY KEY (id);


--
-- Name: saas_plan_quota saas_plan_quota_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plan_quota
    ADD CONSTRAINT saas_plan_quota_pkey PRIMARY KEY (id);


--
-- Name: saas_quota saas_quota_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_quota
    ADD CONSTRAINT saas_quota_pkey PRIMARY KEY (id);


--
-- Name: sys_param sys_param_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_param
    ADD CONSTRAINT sys_param_pkey PRIMARY KEY (id);


--
-- Name: system_param system_param_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.system_param
    ADD CONSTRAINT system_param_pkey PRIMARY KEY (id);


--
-- Name: tenant_dict_item_override tenant_dict_item_override_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_dict_item_override
    ADD CONSTRAINT tenant_dict_item_override_pkey PRIMARY KEY (id);


--
-- Name: tenant_feature_override tenant_feature_override_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_feature_override
    ADD CONSTRAINT tenant_feature_override_pkey PRIMARY KEY (id);


--
-- Name: tenant_menu_override tenant_menu_override_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_menu_override
    ADD CONSTRAINT tenant_menu_override_pkey PRIMARY KEY (id);


--
-- Name: tenant_param_value tenant_param_value_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_param_value
    ADD CONSTRAINT tenant_param_value_pkey PRIMARY KEY (id);


--
-- Name: tenant tenant_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant
    ADD CONSTRAINT tenant_pkey PRIMARY KEY (id);


--
-- Name: tenant_quota_override tenant_quota_override_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_quota_override
    ADD CONSTRAINT tenant_quota_override_pkey PRIMARY KEY (id);


--
-- Name: tenant_quota_usage tenant_quota_usage_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_quota_usage
    ADD CONSTRAINT tenant_quota_usage_pkey PRIMARY KEY (id);


--
-- Name: tenant_subscription tenant_subscription_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_subscription
    ADD CONSTRAINT tenant_subscription_pkey PRIMARY KEY (id);


--
-- Name: user_preference user_preference_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_preference
    ADD CONSTRAINT user_preference_pkey PRIMARY KEY (id);


--
-- Name: user_role user_role_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_role
    ADD CONSTRAINT user_role_pkey PRIMARY KEY (id);


--
-- Name: idx_app_user_account; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_account ON public.app_user USING btree (account);


--
-- Name: idx_app_user_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_company_id ON public.app_user USING btree (company_id);


--
-- Name: idx_app_user_department_department_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_department_department_id ON public.app_user_department USING btree (department_id);


--
-- Name: idx_app_user_department_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_department_id ON public.app_user USING btree (department_id);


--
-- Name: idx_app_user_department_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_department_user_id ON public.app_user_department USING btree (user_id);


--
-- Name: idx_app_user_position_position_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_position_position_id ON public.app_user_position USING btree (position_id);


--
-- Name: idx_app_user_position_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_position_user_id ON public.app_user_position USING btree (user_id);


--
-- Name: idx_app_user_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_user_tenant_id ON public.app_user USING btree (tenant_id);


--
-- Name: idx_audit_log_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_tenant_id ON public.audit_log USING btree (tenant_id);


--
-- Name: idx_audit_log_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_log_user_id ON public.audit_log USING btree (user_id);


--
-- Name: idx_business_unit_org_map_business_unit_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_unit_org_map_business_unit_id ON public.business_unit_org_map USING btree (business_unit_id);


--
-- Name: idx_business_unit_org_map_org_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_unit_org_map_org_id ON public.business_unit_org_map USING btree (org_id);


--
-- Name: idx_business_unit_org_map_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_unit_org_map_tenant_id ON public.business_unit_org_map USING btree (tenant_id);


--
-- Name: idx_business_unit_scope_business_unit_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_unit_scope_business_unit_id ON public.business_unit_scope USING btree (business_unit_id);


--
-- Name: idx_business_unit_scope_role_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_unit_scope_role_permission_id ON public.business_unit_scope USING btree (role_permission_id);


--
-- Name: idx_business_unit_tenant_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_business_unit_tenant_code ON public.business_unit USING btree (tenant_id, code);


--
-- Name: idx_dict_item_dict_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_dict_item_dict_type_id ON public.dict_item USING btree (dict_type_id);


--
-- Name: idx_dict_item_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_dict_item_tenant_id ON public.dict_item USING btree (tenant_id);


--
-- Name: idx_dict_type_tenant_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_dict_type_tenant_code ON public.dict_type USING btree (tenant_id, code);


--
-- Name: idx_login_log_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_login_log_tenant_id ON public.login_log USING btree (tenant_id);


--
-- Name: idx_login_log_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_login_log_user_id ON public.login_log USING btree (user_id);


--
-- Name: idx_org_node_company_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_node_company_id ON public.org_node USING btree (company_id);


--
-- Name: idx_org_node_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_node_parent_id ON public.org_node USING btree (parent_id);


--
-- Name: idx_org_node_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_org_node_tenant_id ON public.org_node USING btree (tenant_id);


--
-- Name: idx_permission_custom_department_department_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_custom_department_department_id ON public.permission_custom_department USING btree (department_id);


--
-- Name: idx_permission_custom_department_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_custom_department_permission_id ON public.permission_custom_department USING btree (permission_id);


--
-- Name: idx_permission_custom_user_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_custom_user_permission_id ON public.permission_custom_user USING btree (permission_id);


--
-- Name: idx_permission_custom_user_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_custom_user_user_id ON public.permission_custom_user USING btree (user_id);


--
-- Name: idx_permission_path; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_path ON public.permission USING btree (path);


--
-- Name: idx_permission_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_permission_tenant_id ON public.permission USING btree (tenant_id);


--
-- Name: idx_plan_feature; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_plan_feature ON public.saas_plan_feature USING btree (plan_id, feature_id);


--
-- Name: idx_plan_quota; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_plan_quota ON public.saas_plan_quota USING btree (plan_id, quota_id);


--
-- Name: idx_position_position_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_position_position_type_id ON public."position" USING btree (position_type_id);


--
-- Name: idx_position_tenant_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_position_tenant_code ON public."position" USING btree (tenant_id, code);


--
-- Name: idx_position_type_tenant_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_position_type_tenant_code ON public.position_type USING btree (tenant_id, code);


--
-- Name: idx_role_permission_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_permission_permission_id ON public.role_permission USING btree (permission_id);


--
-- Name: idx_role_permission_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_permission_role_id ON public.role_permission USING btree (role_id);


--
-- Name: idx_role_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_role_tenant_id ON public.role USING btree (tenant_id);


--
-- Name: idx_saas_feature_feature_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_feature_feature_code ON public.saas_feature USING btree (feature_code);


--
-- Name: idx_saas_plan_plan_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_plan_plan_code ON public.saas_plan USING btree (plan_code);


--
-- Name: idx_saas_quota_quota_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_quota_quota_code ON public.saas_quota USING btree (quota_code);


--
-- Name: idx_sys_param_tenant_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_sys_param_tenant_key ON public.sys_param USING btree (tenant_id, param_key);


--
-- Name: idx_system_param_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_system_param_key ON public.system_param USING btree (param_key);


--
-- Name: idx_tenant_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_code ON public.tenant USING btree (code);


--
-- Name: idx_tenant_dict_item_override_dict_item_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_dict_item_override_dict_item_id ON public.tenant_dict_item_override USING btree (dict_item_id);


--
-- Name: idx_tenant_dict_item_override_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_dict_item_override_tenant_id ON public.tenant_dict_item_override USING btree (tenant_id);


--
-- Name: idx_tenant_feature; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_feature ON public.tenant_feature_override USING btree (tenant_id, feature_id);


--
-- Name: idx_tenant_menu_override_permission_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_menu_override_permission_id ON public.tenant_menu_override USING btree (permission_id);


--
-- Name: idx_tenant_menu_override_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_menu_override_tenant_id ON public.tenant_menu_override USING btree (tenant_id);


--
-- Name: idx_tenant_param_value_param_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_param_value_param_id ON public.tenant_param_value USING btree (param_id);


--
-- Name: idx_tenant_param_value_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_param_value_tenant_id ON public.tenant_param_value USING btree (tenant_id);


--
-- Name: idx_tenant_quota; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_quota ON public.tenant_quota_override USING btree (tenant_id, quota_id);


--
-- Name: idx_tenant_quota_period; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_quota_period ON public.tenant_quota_usage USING btree (tenant_id, quota_code, period_key);


--
-- Name: idx_tenant_subscription_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_subscription_plan_id ON public.tenant_subscription USING btree (plan_id);


--
-- Name: idx_tenant_subscription_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_subscription_tenant_id ON public.tenant_subscription USING btree (tenant_id);


--
-- Name: idx_user_preference_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_preference_user_id ON public.user_preference USING btree (user_id);


--
-- Name: idx_user_role_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_role_role_id ON public.user_role USING btree (role_id);


--
-- Name: idx_user_role_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_role_user_id ON public.user_role USING btree (user_id);


--
-- PostgreSQL database dump complete
--


