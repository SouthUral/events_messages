CREATE OR REPLACE PROCEDURE device.set_messages(IN p_message jsonb)
 LANGUAGE plpgsql
 SECURITY DEFINER
AS $procedure$
    begin
	    INSERT INTO device.messages (
            created_id,
            device_id,
            object_id,
            mes_id,
            mes_time,
            mes_code,
            mes_status,
            mes_data,
            event_value,
            event_data
        )
        VALUES (
            (p_message ->> 'created_id') :: timestamp,
            (p_message ->> 'device_id') :: bigint,
            (p_message ->> 'object_id') :: integer,
            (p_message ->> 'mes_id') :: bigint,
            (p_message ->> 'mes_time') :: timestamp,
            (p_message -> 'event_info' ->> 'code') :: integer,
            (p_message -> 'data' -> 'status_info'),
            (p_message -> 'data'),
            (p_message -> 'event'),
            (p_message -> 'event_data')
        );
    end;
$procedure$
;