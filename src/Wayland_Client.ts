import { File_Descriptor_Claim } from "./File_Descriptor_Claim.ts";
import { Sender } from "./Sender.ts";
import { Send_Message, is_debug_send_message } from "./Send_Message.ts";
import { global_objects, Global_Ids } from "./GlobalObjects.ts";
import { Message_Decoder } from "./Message_Decoder.ts";
import { Wayland_Object } from "./Wayland_Object.ts";
import {
  wl_callback,
  wl_display,
  wl_surface,
  xdg_toplevel,
} from "./protocols/wayland.xml.ts";
import { File_Descriptor, Object_ID, version } from "./wayland_types.ts";
import {
  get_message_and_file_descriptors,
  send_message_and_file_descriptors,
} from "./c_promises.ts";
import { Client_State } from "./c_interop.ts";
import { wayland_debug_time_only } from "./debug.ts" with { type: "macro" };
import {
  Global_ID_To_Object_ID,
  safe_cast_global_id_to_object_id,
} from "./GlobalObjects_To_ID.ts";
import { Role_or_xdg_surface_Object_ID } from "./Role_or_xdg_surface_Object_ID.ts";
import { Surface_Role } from "./Surface_Role.ts";
import { Object_ID_To_Wayland_Object } from "./Object_ID_To_Wayland_Object.ts";

export class Wayland_Client implements File_Descriptor_Claim, Sender {
  drawable_surfaces = new Set<Object_ID<wl_surface>>();

  top_level_surfaces = new Set<Object_ID<xdg_toplevel>>();

  add_frame_draw_request = (callback_id: Object_ID<wl_callback>) => {
    this.frame_draw_requests.push(callback_id);
  };
  /**
   * surface methods
   * */

  get_surface_id_from_role = (
    role_object_id: Role_or_xdg_surface_Object_ID
  ) => {
    return this.roles_to_surfaces.get(role_object_id);
  };

  get_surface_from_role = (role_object_id: Role_or_xdg_surface_Object_ID) => {
    const surface_id = this.get_surface_id_from_role(role_object_id);
    if (surface_id === undefined) {
      return undefined;
    }
    return this.get_object(surface_id)?.delegate;
  };

  // get_role_data_from_role = <T extends Surface_Role["type"]>(
  //   role_object_id: Object_ID,
  //   role: T
  // ): Extract<Surface_Role, { type: T }>["data"] | null => {
  //   const surface = this.get_surface_from_role(role_object_id);
  //   if (!surface) {
  //     return null;
  //   }
  //   if (!surface.role) {
  //     return null;
  //   }
  //   if (surface.role.type !== role) {
  //     return null;
  //   }
  //   return surface.role.data as any;
  // };

  register_role_to_surface = (
    role_id: Role_or_xdg_surface_Object_ID,
    surface_id: Object_ID<wl_surface>
  ) => {
    this.roles_to_surfaces.set(role_id, surface_id);
  };

  unregister_role_to_surface = (role_id: Role_or_xdg_surface_Object_ID) => {
    this.roles_to_surfaces.delete(role_id);
  };

  /**
   * Seed if maybe_desceneding_id is a descendant of surface_id
   * @param s
   * @param surface_id
   * @param maybe_descendant_id
   */
  find_descendant_surface = (
    surface_id: Object_ID<wl_surface>,
    maybe_descendant_id: Object_ID<wl_surface>
  ): boolean => {
    const surface = this.get_object(surface_id)?.delegate;
    if (!surface) {
      return false;
    }
    for (const child_id of surface.children_in_draw_order) {
      if (child_id === maybe_descendant_id) {
        return true;
      }
    }

    for (const child_id of surface.children_in_draw_order) {
      if (child_id === null) {
        continue;
      }
      if (this.find_descendant_surface(child_id, maybe_descendant_id)) {
        return true;
      }
    }
    return false;
  };

  compositor_version: version = 1;

  /**
   * A map from a role_object_id to a surface_id
   */
  roles_to_surfaces: Map<Role_or_xdg_surface_Object_ID, Object_ID<wl_surface>> =
    new Map();

  frame_draw_requests: Object_ID<wl_callback>[] = [];

  // object_state: Object_State = {};

  send_error = (object_id: Object_ID, code: number, message: string) => {
    wl_display.error(
      this,
      safe_cast_global_id_to_object_id(Global_Ids.wl_display),
      object_id,
      code,
      message
    );
  };

  message_buffer = new Uint8Array(1024);
  file_descriptor_buffer = new Uint32Array(255);

  send_message_buffer = new Uint8Array(1024);
  send_file_descriptor_buffer = new Uint32Array(255);

  objects: Map<Object_ID, Wayland_Object<any>> = new Map();
  _global_binds: Map<Global_Ids, Map<Object_ID, version>> = new Map();

  get_global_binds = <T extends Global_Ids>(
    global_id: T
  ): Map<Global_ID_To_Object_ID<T>, version> | undefined => {
    return this._global_binds.get(global_id) as Map<
      Global_ID_To_Object_ID<T>,
      version
    >;
  };

  remove_global_bind = <T extends Global_Ids>(
    global_id: T,
    object_id: Global_ID_To_Object_ID<T>
  ) => {
    const set = this._global_binds.get(global_id);
    if (set == undefined) {
      return;
    }
    set.delete(object_id);
  };

  /**
   * Add a bound object_id to a list
   * of global_ids. SO that you can
   * ask, What are all the objects bound
   * to this global for this client?
   * @param global_id
   * @param object_id
   */
  add_global_bind = <T extends Global_Ids>(
    global_id: T,
    object_id: Global_ID_To_Object_ID<T>,
    version: version
  ) => {
    if (!this._global_binds.has(global_id)) {
      this._global_binds.set(global_id, new Map());
    }
    this._global_binds.get(global_id)?.set(object_id, version);
  };

  add_object = <T extends Object_ID>(
    object_id: T,
    object: T extends Object_ID<infer K> ? K : never
  ) => {
    if (object == undefined) {
      console.log("object is undefined cannot add it", object_id);
    }
    if (this.objects.has(object_id)) {
      console.log("object already exists", object_id);
    }
    this.objects.set(object_id, object);
  };
  remove_object = (object_id: Object_ID) => {
    this.objects.delete(object_id);
  };

  get_object = <T extends Object_ID>(
    object_id: T
  ): Object_ID_To_Wayland_Object<T> | undefined => {
    return (
      this.objects.get(object_id) ?? (global_objects.objects[object_id] as any)
    );
  };

  // get_object_delegate_cast_to = <T>(object_id: Object_ID): T | undefined => {
  //   return (
  //     (this.get_object(object_id)?.delegate as T) ??
  //     (global_objects.objects[object_id]?.delegate as T)
  //   );
  // };
  pending_message: Send_Message[] = [];

  message_decoder = new Message_Decoder();

  unclaimed_file_descriptors: File_Descriptor[] = [];

  constructor(
    public client_socket: number,
    public client_state: Client_State
  ) {
    if (wayland_debug_time_only()) {
      console.log(`new client`, client_socket);
    }
  }

  main_loop = async () => {
    while (true) {
      for (const message of this.pending_message) {
        const should_continue = await this.send_pending_messages(message);
        if (!should_continue) {
          return;
        }
      }

      this.pending_message = [];

      const message = await get_message_and_file_descriptors(
        this.client_socket,
        this.message_buffer,
        this.file_descriptor_buffer
      );
      const should_continue = this.parse_messages(message);
      if (!should_continue) {
        return;
      }
    }
  };

  /**
   *
   * Adds the message to the pending message queue
   */
  send = (data: Send_Message) => {
    this.pending_message.push(data);
  };
  /**
   *
   * @param message
   * @returns Returns if we should continue listening or sending on this socket any more
   * returns falsy mostly if the client has disconnected
   */
  send_pending_messages = async (message: Send_Message): Promise<boolean> => {
    if (wayland_debug_time_only()) {
      if (is_debug_send_message(message)) {
        console.log(
          `client#${this.client_socket} -> ${message.object_name}@${message.object_id}.${message.message_name}(${message.message_args.map(({ signature, value }) => `${signature} = ${value}`).join(", ")})`
        );
      }
    }
    /**
     * 8 bytes is the header length + the length of the message
     * #### Header is
     * - 4 bytes for object_id
     * - 2 bytes for opcode
     * - 2 bytes for size
     */
    const length = 8 + message.data.length;

    this.send_message_buffer.set(
      new Uint8Array([
        message.object_id & 0xff,
        (message.object_id >> 8) & 0xff,
        (message.object_id >> 16) & 0xff,
        (message.object_id >> 24) & 0xff,
        message.opcode & 0xff,
        (message.opcode >> 8) & 0xff,
        length & 0xff,
        (length >> 8) & 0xff,
      ])
    );
    if (message.data.length > 0) {
      this.send_message_buffer.set(message.data, 8);
    }
    this.send_file_descriptor_buffer[0] = message.file_descriptor ?? 0;

    let offset = 0;
    while (true) {
      if (offset > 0) {
        console.log("\nsending more data\n");
      }
      const data_view = this.send_message_buffer.subarray(offset, length);
      /**
       * only send the file descriptor on the first message
       */
      const file_descriptor_view =
        offset == 0 && message.file_descriptor != null
          ? this.send_file_descriptor_buffer.subarray(0, 1)
          : this.send_file_descriptor_buffer.subarray(0, 0);

      this.send_message_buffer.subarray();

      const { should_continue, bytes_written } =
        await send_message_and_file_descriptors(
          this.client_socket,
          data_view,
          file_descriptor_view
        );

      if (!should_continue) {
        return false;
      }

      offset += bytes_written;
      if (offset >= length) {
        break;
      }
    }
    return true;
  };

  /**
   *
   * @param param0
   * @returns true if should continue listening for more messages
   */
  parse_messages = ({
    should_continue,
    bytes_read,
    number_of_file_descriptors: file_descriptors,
  }: {
    should_continue: boolean;
    bytes_read: number;
    number_of_file_descriptors: number;
  }): boolean => {
    for (let i = 0; i < file_descriptors; i++) {
      if (wayland_debug_time_only()) {
        console.log(
          `client#${this.client_socket}: Received File descriptor`,
          this.file_descriptor_buffer[i]
        );
      }
      this.unclaimed_file_descriptors.push(
        this.file_descriptor_buffer[i] as File_Descriptor
      );
    }

    if (bytes_read < 0) {
      console.log("error reading message from client");
      console.error("Failed to read message from client");
      return false;
    }
    if (!should_continue) {
      return false;
    }

    if (bytes_read === 0) {
      /**
       * Time out
       */
      return true;
    }

    const new_messages = this.message_decoder.consume(
      this.message_buffer,
      bytes_read
    );

    for (const message of new_messages) {
      const object = this.get_object(message.object_id);
      if (object == undefined) {
        console.log("can not do request on undefined", message.object_id);
      }
      object?.on_request(this, message);
    }

    return true;
  };

  claim_file_descriptor: File_Descriptor_Claim["claim_file_descriptor"] =
    () => {
      const number = this.unclaimed_file_descriptors.shift();
      if (number == undefined) {
        return null as unknown as File_Descriptor;
      }
      return number;
    };
}
